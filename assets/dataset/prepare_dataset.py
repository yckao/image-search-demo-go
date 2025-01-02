import os
import mimetypes
import time
import statistics
from concurrent.futures import ThreadPoolExecutor
from http.client import HTTPConnection
from urllib.parse import urlparse


class UploadStats:
    def __init__(self):
        self.latencies = []
        self.success_count = 0
        self.failed_count = 0
        self.start_time = None
        self.end_time = None

    def add_latency(self, latency):
        self.latencies.append(latency)

    def print_summary(self):
        if not self.latencies:
            print("No requests were made.")
            return

        print("\n=== Upload Statistics ===")
        print(f"Total requests: {len(self.latencies)}")
        print(f"Successful requests: {self.success_count}")
        print(f"Failed requests: {self.failed_count}")
        print(f"Total time: {(self.end_time - self.start_time):.2f} seconds")
        print("\nLatency Statistics (seconds):")
        print(f"Average: {statistics.mean(self.latencies):.3f}")
        print(f"Median: {statistics.median(self.latencies):.3f}")
        print(f"Min: {min(self.latencies):.3f}")
        print(f"Max: {max(self.latencies):.3f}")
        if len(self.latencies) > 1:
            print(f"Std Dev: {statistics.stdev(self.latencies):.3f}")


def create_multipart_form(file_path):
    boundary = "----WebKitFormBoundary7MA4YWxkTrZu0gW"

    content_type = mimetypes.guess_type(file_path)[0] or "application/octet-stream"

    with open(file_path, "rb") as f:
        file_content = f.read()

    body = []
    body.append(f"--{boundary}".encode())
    body.append(f'Content-Disposition: form-data; name="file"; filename="{os.path.basename(file_path)}"'.encode())
    body.append(f"Content-Type: {content_type}".encode())
    body.append(b"")
    body.append(file_content)
    body.append(f"--{boundary}--".encode())

    return b"\r\n".join(body), boundary


def upload_file(file_path, url, stats):
    try:
        start_time = time.time()

        parsed_url = urlparse(url)
        body, boundary = create_multipart_form(file_path)
        headers = {
            "Content-Type": f"multipart/form-data; boundary={boundary}",
            "Accept": "application/json",
        }

        conn = HTTPConnection(parsed_url.netloc)
        conn.request("POST", parsed_url.path, body=body, headers=headers)

        response = conn.getresponse()
        end_time = time.time()
        latency = end_time - start_time

        status = response.status
        filename = os.path.basename(file_path)

        # Record statistics
        stats.add_latency(latency)
        if 200 <= status < 300:
            stats.success_count += 1
        else:
            stats.failed_count += 1

        print(f"File: {filename} - Status: {status} - Latency: {latency:.3f}s")

        conn.close()
        return status
    # pylint: disable-next=broad-exception-caught
    except Exception as e:
        stats.failed_count += 1
        print(f"Error uploading {file_path}: {str(e)}")
        return None


def main():
    directory_path = os.getenv("DATASET_PATH") or "/app/val2014"
    url = os.getenv("UPLOAD_URL") or "http://localhost:8080/images"
    max_workers = 5

    # Initialize statistics
    stats = UploadStats()

    # Get all files in directory
    files = [
        os.path.join(directory_path, f)
        for f in os.listdir(directory_path)
        if os.path.isfile(os.path.join(directory_path, f))
    ]

    print(f"Starting upload of {len(files)} files...")
    stats.start_time = time.time()

    # Upload files concurrently
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = [executor.submit(upload_file, file_path, url, stats) for file_path in files]

        # Wait for all uploads to complete
        for future in futures:
            future.result()

    stats.end_time = time.time()
    stats.print_summary()


if __name__ == "__main__":
    main()
