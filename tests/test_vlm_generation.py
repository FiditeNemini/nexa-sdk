# This test case runs successfully locally but unexpectedly crashes during GitHub CI/CD.
# It has been temporarily disabled. Please run this test manually before submitting a pull request.

# from nexa.gguf import NexaVLMInference
# from nexa.gguf.lib_utils import is_gpu_available
# from tempfile import TemporaryDirectory
# import os
# import subprocess


# def download_model(url, output_dir):
#     """
#     Download a file from a given URL using curl, if it doesn't already exist.

#     Args:
#     - url: str, the URL of the file to download.
#     - output_dir: str, the directory where the file should be saved.

#     Returns:
#     - str: The path to the downloaded file.
#     """
#     file_name = url.split("/")[-1]
#     output_path = os.path.join(output_dir, file_name)

#     if os.path.exists(output_path):
#         print(
#             f"File {file_name} already exists in {output_dir}. Skipping download.")
#         return output_path

#     try:
#         subprocess.run(["curl", url, "--output", output_path], check=True)
#         print(f"Downloaded {file_name} to {output_dir}")
#     except subprocess.CalledProcessError as e:
#         print(f"Failed to download {file_name}: {e}")
#         raise

#     return output_path


# # Initialize the model
# model = NexaVLMInference(
#     model_path="nanollava",
#     verbose=False,
#     n_gpu_layers=-1 if is_gpu_available() else 0,
# )


# # Test VLM generation without an image
# def test_text_only_generation():
#     output = model.create_chat_completion(
#         messages=[
#             {"role": "user", "content": "What is the capital of France?"}
#         ],
#         max_tokens=100,
#         stream=False
#     )
#     response = output["choices"][0]["message"]["content"]
#     print("Text-only response:", response)
#     assert "Paris" in response, "The response should mention Paris"


# # Test VLM generation with an image
# def test_image_description():
#     img_url = "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
#     with TemporaryDirectory() as temp_dir:
#         img_path = download_model(img_url, temp_dir)
#         user_input = "Describe this image in detail."
#         output = model._chat(user_input, img_path)
#         response = ""
#         for chunk in output:
#             delta = chunk["choices"][0]["delta"]
#             if "content" in delta:
#                 response += delta["content"]
#         print("Image description:", response)
#         assert len(response) > 50, "The image description should be detailed"


# # Test streaming output
# def test_streaming_output():
#     global model
#     messages = [
#         {"role": "user", "content": "Write a short story about a robot learning to paint."}
#     ]
#     output = model.create_chat_completion(
#         messages=messages, max_tokens=200, stream=True)
#     story = ""
#     for chunk in output:
#         if "choices" in chunk and len(chunk["choices"]) > 0:
#             delta = chunk["choices"][0]["delta"]
#             if "content" in delta:
#                 story += delta["content"]
#                 print(delta["content"], end="", flush=True)
#     print("\nFull story:", story)
#     assert len(story) > 100, "The generated story should be of substantial length"


# # Main execution
# if __name__ == "__main__":
#     print("=== Testing Text-Only Generation ===")
#     test_text_only_generation()

#     print("\n=== Testing Image Generation ===")
#     test_image_description()

#     print("\n=== Testing Streaming Output ===")
#     test_streaming_output()
