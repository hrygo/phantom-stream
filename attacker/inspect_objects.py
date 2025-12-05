import re
import sys

def inspect_objects(file_path, object_ids):
    try:
        with open(file_path, 'rb') as f:
            content = f.read()
            content_str = content.decode('latin-1')

        print(f"Inspecting objects {object_ids} in {file_path}...")

        for obj_id in object_ids:
            # Find the object definition
            pattern = re.compile(r'({}\s+0\s+obj.*?endobj)'.format(obj_id), re.DOTALL)
            match = pattern.search(content_str)
            
            if match:
                obj_content = match.group(1)
                print(f"\n--- Object {obj_id} ---")
                
                # Print the dictionary part (before stream)
                dict_part = obj_content.split('stream')[0]
                print(f"Dictionary: {dict_part.strip()}")
                
                # Check for stream
                if 'stream' in obj_content:
                    stream_match = re.search(r'stream(.*?)endstream', obj_content, re.DOTALL)
                    if stream_match:
                        stream_data = stream_match.group(1)
                        print(f"Stream Length: {len(stream_data)} bytes")
                        # Print first few bytes in hex
                        preview = stream_data[:50].encode('latin-1').hex()
                        print(f"Stream Preview: {preview}...")
                    else:
                        print("Stream detected but regex failed to extract.")
                else:
                    print("No stream found.")
            else:
                print(f"Object {obj_id} not found.")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python3 inspect_objects.py <pdf_file> <obj_id1> [obj_id2 ...]")
        sys.exit(1)
    
    file_path = sys.argv[1]
    ids = sys.argv[2:]
    inspect_objects(file_path, ids)
