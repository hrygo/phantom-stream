import zlib
import re
import sys

def analyze_stream(file_path, obj_id):
    try:
        with open(file_path, 'rb') as f:
            content = f.read()
            content_str = content.decode('latin-1')

        pattern = re.compile(r'{}\s+0\s+obj[\s\S]*?>>\s*stream\s*([\s\S]*?)\s*endstream'.format(obj_id), re.DOTALL)
        match = pattern.search(content_str)
        
        if not match:
            print(f"Object {obj_id} stream not found")
            return

        stream_data = match.group(1).encode('latin-1')
        print(f"Object {obj_id} Compressed Size: {len(stream_data)}")

        try:
            decompressed = zlib.decompress(stream_data)
            print(f"Decompressed Size: {len(decompressed)}")
            print(f"First 20 bytes: {decompressed[:20].hex()}")
        except Exception as e:
            print(f"Decompression failed: {e}")
            # Try stripping whitespace?
            try:
                s = stream_data.strip()
                decompressed = zlib.decompress(s)
                print(f"Decompressed Size (stripped): {len(decompressed)}")
            except:
                pass

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    analyze_stream(sys.argv[1], sys.argv[2])
