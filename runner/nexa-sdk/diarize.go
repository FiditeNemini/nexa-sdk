package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"
import (
	"log/slog"
	"unsafe"
)

// DiarizeConfig represents diarization processing configuration
type DiarizeConfig struct {
	MinSpeakers int32
	MaxSpeakers int32
}

func (dc DiarizeConfig) toCPtr() *C.ml_DiarizeConfig {
	cPtr := (*C.ml_DiarizeConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_DiarizeConfig{}))))
	*cPtr = C.ml_DiarizeConfig{}

	cPtr.min_speakers = C.int32_t(dc.MinSpeakers)
	cPtr.max_speakers = C.int32_t(dc.MaxSpeakers)

	return cPtr
}

func freeDiarizeConfig(cPtr *C.ml_DiarizeConfig) {
	if cPtr != nil {
		C.free(unsafe.Pointer(cPtr))
	}
}

// DiarizeSpeechSegment represents a speech segment with speaker label
type DiarizeSpeechSegment struct {
	StartTime    float32
	EndTime      float32
	SpeakerLabel string
}

func newDiarizeSpeechSegmentFromCPtr(c *C.ml_DiarizeSpeechSegment) DiarizeSpeechSegment {
	segment := DiarizeSpeechSegment{}

	if c == nil {
		return segment
	}

	segment.StartTime = float32(c.start_time)
	segment.EndTime = float32(c.end_time)

	if c.speaker_label != nil {
		segment.SpeakerLabel = C.GoString(c.speaker_label)
	}

	return segment
}

func freeDiarizeSpeechSegment(ptr *C.ml_DiarizeSpeechSegment) {
	if ptr == nil {
		return
	}
	if ptr.speaker_label != nil {
		mlFree(unsafe.Pointer(ptr.speaker_label))
	}
}

// DiarizeModelConfig represents diarization model configuration
type DiarizeModelConfig struct {
	NCtx               int32
	NThreads           int32
	NThreadsBatch      int32
	NBatch             int32
	NUbatch            int32
	NSeqMax            int32
	NGpuLayers         int32
	ChatTemplatePath   string
	ChatTemplateContent string
	EnableSampling     bool
	GrammarStr         string
	MaxTokens          int32
	EnableThinking     bool
	Verbose            bool
	QnnModelFolderPath string
	QnnLibFolderPath   string
}

func (mc DiarizeModelConfig) toCPtr() *C.ml_ModelConfig {
	cPtr := (*C.ml_ModelConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_ModelConfig{}))))
	*cPtr = C.ml_ModelConfig{}

	cPtr.n_ctx = C.int32_t(mc.NCtx)
	cPtr.n_threads = C.int32_t(mc.NThreads)
	cPtr.n_threads_batch = C.int32_t(mc.NThreadsBatch)
	cPtr.n_batch = C.int32_t(mc.NBatch)
	cPtr.n_ubatch = C.int32_t(mc.NUbatch)
	cPtr.n_seq_max = C.int32_t(mc.NSeqMax)
	cPtr.n_gpu_layers = C.int32_t(mc.NGpuLayers)

	if mc.ChatTemplatePath != "" {
		cPtr.chat_template_path = C.CString(mc.ChatTemplatePath)
	}
	if mc.ChatTemplateContent != "" {
		cPtr.chat_template_content = C.CString(mc.ChatTemplateContent)
	}
	cPtr.enable_sampling = C.bool(mc.EnableSampling)
	if mc.GrammarStr != "" {
		cPtr.grammar_str = C.CString(mc.GrammarStr)
	}
	cPtr.max_tokens = C.int32_t(mc.MaxTokens)
	cPtr.enable_thinking = C.bool(mc.EnableThinking)
	cPtr.verbose = C.bool(mc.Verbose)

	if mc.QnnModelFolderPath != "" {
		cPtr.qnn_model_folder_path = C.CString(mc.QnnModelFolderPath)
	}
	if mc.QnnLibFolderPath != "" {
		cPtr.qnn_lib_folder_path = C.CString(mc.QnnLibFolderPath)
	}

	return cPtr
}

// DiarizeCreateInput represents input parameters for diarization creation
type DiarizeCreateInput struct {
	ModelName  string
	ModelPath  string
	Config     DiarizeModelConfig
	PluginID   string
	DeviceID   string
	LicenseID  string
	LicenseKey string
}

func (dci DiarizeCreateInput) toCPtr() *C.ml_DiarizeCreateInput {
	cPtr := (*C.ml_DiarizeCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_DiarizeCreateInput{}))))
	*cPtr = C.ml_DiarizeCreateInput{}

	if dci.ModelName != "" {
		cPtr.model_name = C.CString(dci.ModelName)
	}
	if dci.ModelPath != "" {
		cPtr.model_path = C.CString(dci.ModelPath)
	}
	cPtr.config = *dci.Config.toCPtr()
	if dci.PluginID != "" {
		cPtr.plugin_id = C.CString(dci.PluginID)
	}
	if dci.DeviceID != "" {
		cPtr.device_id = C.CString(dci.DeviceID)
	}
	if dci.LicenseID != "" {
		cPtr.license_id = C.CString(dci.LicenseID)
	}
	if dci.LicenseKey != "" {
		cPtr.license_key = C.CString(dci.LicenseKey)
	}

	return cPtr
}

func freeDiarizeCreateInput(cPtr *C.ml_DiarizeCreateInput) {
	if cPtr != nil {
		if cPtr.model_name != nil {
			C.free(unsafe.Pointer(cPtr.model_name))
		}
		if cPtr.model_path != nil {
			C.free(unsafe.Pointer(cPtr.model_path))
		}
		// config is a struct, not a pointer, so no need to free it
		if cPtr.plugin_id != nil {
			C.free(unsafe.Pointer(cPtr.plugin_id))
		}
		if cPtr.device_id != nil {
			C.free(unsafe.Pointer(cPtr.device_id))
		}
		if cPtr.license_id != nil {
			C.free(unsafe.Pointer(cPtr.license_id))
		}
		if cPtr.license_key != nil {
			C.free(unsafe.Pointer(cPtr.license_key))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// DiarizeInferInput represents input parameters for diarization inference
type DiarizeInferInput struct {
	AudioPath string
	Config    *DiarizeConfig
}

func (dii DiarizeInferInput) toCPtr() *C.ml_DiarizeInferInput {
	cPtr := (*C.ml_DiarizeInferInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_DiarizeInferInput{}))))
	*cPtr = C.ml_DiarizeInferInput{}

	cPtr.audio_path = C.CString(dii.AudioPath)
	if dii.Config != nil {
		cPtr.config = dii.Config.toCPtr()
	} else {
		cPtr.config = nil
	}

	return cPtr
}

func freeDiarizeInferInput(cPtr *C.ml_DiarizeInferInput) {
	if cPtr != nil {
		if cPtr.audio_path != nil {
			C.free(unsafe.Pointer(cPtr.audio_path))
		}
		if cPtr.config != nil {
			freeDiarizeConfig(cPtr.config)
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// DiarizeInferOutput represents output from diarization inference
type DiarizeInferOutput struct {
	Segments     []DiarizeSpeechSegment
	NumSpeakers  int32
	Duration     float32
	ProfileData  ProfileData
}

func newDiarizeInferOutputFromCPtr(c *C.ml_DiarizeInferOutput) DiarizeInferOutput {
	output := DiarizeInferOutput{}

	if c == nil {
		return output
	}

	// Convert segments array
	if c.segments != nil && c.segment_count > 0 {
		segments := unsafe.Slice((*C.ml_DiarizeSpeechSegment)(unsafe.Pointer(c.segments)), int(c.segment_count))
		output.Segments = make([]DiarizeSpeechSegment, c.segment_count)
		for i := range output.Segments {
			output.Segments[i] = newDiarizeSpeechSegmentFromCPtr(&segments[i])
		}
	}

	output.NumSpeakers = int32(c.num_speakers)
	output.Duration = float32(c.duration)
	output.ProfileData = newProfileDataFromCPtr(c.profile_data)

	return output
}

func freeDiarizeInferOutput(ptr *C.ml_DiarizeInferOutput) {
	if ptr == nil {
		return
	}
	if ptr.segments != nil && ptr.segment_count > 0 {
		segments := unsafe.Slice((*C.ml_DiarizeSpeechSegment)(unsafe.Pointer(ptr.segments)), int(ptr.segment_count))
		for i := range segments {
			freeDiarizeSpeechSegment(&segments[i])
		}
		mlFree(unsafe.Pointer(ptr.segments))
	}
}

// Diarize represents a diarization instance
type Diarize struct {
	ptr *C.ml_Diarize
}

// NewDiarize creates a new diarization instance
func NewDiarize(input DiarizeCreateInput) (*Diarize, error) {
	slog.Debug("NewDiarize called", "input", input)

	cInput := input.toCPtr()
	defer freeDiarizeCreateInput(cInput)

	var cHandle *C.ml_Diarize
	res := C.ml_diarize_create(cInput, &cHandle)
	if res < 0 {
		return nil, SDKError(res)
	}

	return &Diarize{ptr: cHandle}, nil
}

// Destroy destroys the diarization instance and frees associated resources
func (d *Diarize) Destroy() error {
	slog.Debug("Destroy called", "ptr", d.ptr)

	if d.ptr == nil {
		return nil
	}

	res := C.ml_diarize_destroy(d.ptr)
	if res < 0 {
		return SDKError(res)
	}
	d.ptr = nil
	return nil
}

// Infer performs speaker diarization on audio file
func (d *Diarize) Infer(input DiarizeInferInput) (DiarizeInferOutput, error) {
	slog.Debug("Infer called", "input", input)

	cInput := input.toCPtr()
	defer freeDiarizeInferInput(cInput)

	var cOutput C.ml_DiarizeInferOutput
	defer freeDiarizeInferOutput(&cOutput)

	res := C.ml_diarize_infer(d.ptr, cInput, &cOutput)
	if res < 0 {
		return DiarizeInferOutput{}, SDKError(res)
	}

	output := newDiarizeInferOutputFromCPtr(&cOutput)
	return output, nil
}

