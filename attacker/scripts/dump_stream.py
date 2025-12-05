import zlib
import re
import sys

def dump_stream(file_path, obj_id):
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
        
        try:
            decompressed = zlib.decompress(stream_data)
            print(f"--- Object {obj_id} Decompressed Content ---")
            print(decompressed.decode('latin-1', errors='ignore'))
            print("-------------------------------------------")
        except Exception as e:
            print(f"Decompression failed: {e}")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    dump_stream(sys.argv[1], sys.argv[2])
