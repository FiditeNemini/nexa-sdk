package service

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
)

// KeepAliveGet retrieves a model from the keepalive cache or creates it if not found
// This avoids the overhead of repeatedly loading/unloading models from disk
func KeepAliveGet[T any](name string, param types.ModelParam, reset bool) (*T, error) {
	t, err := keepAliveGet[T](name, param, reset)
	if err != nil {
		return nil, err
	}
	return t.(*T), nil
}

var keepAlive keepAliveService

// current only support keepalive one model
type keepAliveService struct {
	models map[string]*modelKeepInfo // Map of model name to model info
	stopCh chan struct{}             // Channel to stop the cleanup goroutine

	sync.Mutex // Protects concurrent access to models map
}

// modelKeepInfo holds metadata for a cached model instance
type modelKeepInfo struct {
	model    keepable
	param    types.ModelParam
	lastTime time.Time
}

// keepable interface defines objects that can be managed by the keepalive service
// Objects must support cleanup and reset operations
type keepable interface {
	Destroy() error
}

type keepResetable interface {
	keepable
	Reset() error
}

// start begins the background cleanup process that removes unused models
// Runs a ticker every 5 seconds to check for models that exceed the keepalive timeout
func (keepAlive *keepAliveService) start() {
	keepAlive.models = make(map[string]*modelKeepInfo)
	keepAlive.stopCh = make(chan struct{})

	go func() {
		t := time.NewTicker(5 * time.Second)
		for {
			select {
			// Stop signal received - cleanup all models and exit
			case <-keepAlive.stopCh:
				keepAlive.Lock()
				for name, model := range keepAlive.models {
					model.model.Destroy()
					delete(keepAlive.models, name)
				}
				keepAlive.Unlock()
				return

			// Periodic cleanup - remove models that haven't been used recently
			case <-t.C:
				keepAlive.Lock()
				for name, model := range keepAlive.models {
					if time.Since(model.lastTime).Milliseconds()/1000 > config.Get().KeepAlive {
						model.model.Destroy()
						delete(keepAlive.models, name)
					}
				}
				keepAlive.Unlock()
			}
		}
	}()
}

// keepAliveGet retrieves a cached model or creates a new one if not found
// Ensures only one model is kept in memory at a time by clearing others
func keepAliveGet[T any](name string, param types.ModelParam, reset bool) (any, error) {
	keepAlive.Lock()
	defer keepAlive.Unlock()

	name, quant := utils.NormalizeModelName(name)
	slog.Debug("KeepAliveGet", "name", name, "quant", quant, "param", param)

	s := store.Get()

	manifest, err := s.GetManifest(name)
	if err != nil {
		return nil, err
	}

	var modelfile string
	if quant != "" {
		if fileinfo, exists := manifest.ModelFile[quant]; !exists {
			return nil, fmt.Errorf("quantization %s not found for model %s", quant, name)
		} else if !fileinfo.Downloaded {
			return nil, fmt.Errorf("quantization %s not downloaded for model %s", quant, name)
		} else {
			modelfile = s.ModelfilePath(manifest.Name, fileinfo.Name)
		}
	} else {
		// fallback to first downloaded model file
		quants := make([]string, 0, len(manifest.ModelFile))
		for quant, v := range manifest.ModelFile {
			if v.Downloaded {
				quants = append(quants, quant)
				break
			}
		}
		quant = slices.Min(quants)
		slog.Debug("KeepAliveGet quant fallback", "quant", quant)
		modelfile = s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name) // at least one is downloaded
	}

	// Check if model already exists in cache
	model, ok := keepAlive.models[name+":"+quant]
	if ok && reflect.DeepEqual(model.param, param) {
		if reset {
			if r, ok := model.model.(keepResetable); ok {
				r.Reset()
			}
		}
		model.lastTime = time.Now()
		return model.model, nil
	}

	// Clear existing models to ensure only one is in memory
	// This prevents memory overflow but limits to single model usage
	// TODO: unload model due to free ram/vram
	for name, model := range keepAlive.models {
		model.model.Destroy()
		delete(keepAlive.models, name)
	}

	var t keepable
	var e error
	switch reflect.TypeFor[T]() {
	case reflect.TypeFor[nexa_sdk.LLM]():
		t, e = nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			Config: nexa_sdk.ModelConfig{
				NCtx:         param.NCtx,
				NGpuLayers:   param.NGpuLayers,
				SystemPrompt: param.SystemPrompt,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.VLM]():
		var mmproj string
		if manifest.MMProjFile.Name != "" {
			mmproj = s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
		}
		var tokenizer string
		if manifest.TokenizerFile.Name != "" {
			tokenizer = s.ModelfilePath(manifest.Name, manifest.TokenizerFile.Name)
		}
		t, e = nexa_sdk.NewVLM(nexa_sdk.VlmCreateInput{
			ModelName:     manifest.ModelName,
			ModelPath:     modelfile,
			MmprojPath:    mmproj,
			TokenizerPath: tokenizer,
			Config: nexa_sdk.ModelConfig{
				NCtx:         param.NCtx,
				NGpuLayers:   param.NGpuLayers,
				SystemPrompt: param.SystemPrompt,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.Embedder]():
		t, e = nexa_sdk.NewEmbedder(nexa_sdk.EmbedderCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			PluginID:  manifest.PluginId,
			DeviceID:  manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.Reranker]():
		t, e = nexa_sdk.NewReranker(nexa_sdk.RerankerCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			PluginID:  manifest.PluginId,
			DeviceID:  manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.TTS]():
		t, e = nexa_sdk.NewTTS(nexa_sdk.TtsCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			PluginID:  manifest.PluginId,
			DeviceID:  manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.ASR]():
		t, e = nexa_sdk.NewASR(nexa_sdk.AsrCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			Config: nexa_sdk.ASRModelConfig{
				NCtx:       param.NCtx,
				NGpuLayers: param.NGpuLayers,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.Diarize]():
		t, e = nexa_sdk.NewDiarize(nexa_sdk.DiarizeCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelfile,
			Config: nexa_sdk.DiarizeModelConfig{
				NCtx:       param.NCtx,
				NGpuLayers: param.NGpuLayers,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.CV]():
		t, e = nexa_sdk.NewCV(nexa_sdk.CVCreateInput{
			ModelName: manifest.ModelName,
			Config: nexa_sdk.CVModelConfig{
				DetModelPath: modelfile,
				RecModelPath: modelfile,
			},
			PluginID: manifest.PluginId,
			DeviceID: manifest.DeviceId,
		})
	case reflect.TypeFor[nexa_sdk.ImageGen]():
		// For image generation models, use the model directory path instead of specific file
		modelDir := s.ModelfilePath(manifest.Name, "")
		t, e = nexa_sdk.NewImageGen(nexa_sdk.ImageGenCreateInput{
			ModelName: manifest.ModelName,
			ModelPath: modelDir,
			PluginID:  manifest.PluginId,
			DeviceID:  manifest.DeviceId,
		})
	default:
		panic(fmt.Sprintf("not support type: %+#v", t))
	}
	if e != nil {
		return nil, e
	}
	model = &modelKeepInfo{
		model:    t,
		param:    param,
		lastTime: time.Now(),
	}
	keepAlive.models[name+":"+quant] = model

	return model.model, nil
}

// stop signals the cleanup goroutine to terminate
func (keepAlive *keepAliveService) stop() {
	keepAlive.stopCh <- struct{}{}
}
