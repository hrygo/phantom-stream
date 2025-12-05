import re
import zlib
import sys

def analyze_all_streams(file_path):
    try:
        with open(file_path, 'rb') as f:
            content = f.read()
            content_str = content.decode('latin-1')

        print(f"[*] Scanning all streams in {file_path}...")

        # Regex to find all streams
        # Format: ID 0 obj ... stream ... endstream
        # Captures: ID, Dictionary, Stream Content
        stream_pattern = re.compile(r'(\d+)\s+0\s+obj(.*?)stream([\s\S]*?)endstream', re.DOTALL)
        
        count = 0
        hits = 0
        
        for match in stream_pattern.finditer(content_str):
            count += 1
            obj_id = match.group(1)
            # dict_content = match.group(2) # Unused for now
            stream_content = match.group(3)
            
            # Strip leading CRLF usually present after 'stream'
            stream_data = stream_content.strip('\r\n').encode('latin-1')
            
            try:
                # Try decompressing
                decompressed = zlib.decompress(stream_data)
                text = decompressed.decode('latin-1', errors='ignore')
                
                found = False
                if "b78b" in text:
                    print(f"[!] Found 'b78b' in Object {obj_id}")
                    found = True
                
                if found:
                    hits += 1
                    # Print some context
                    idx = text.find("b78b")
                    start = max(0, idx - 20)
                    end = min(len(text), idx + 50)
                    print(f"    Context: ...{text[start:end].replace(chr(10), ' ')}...")

            except Exception:
                # Not zlib compressed or corrupt
                pass
        
        print(f"[*] Scanned {count} streams. Found {hits} hits.")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 analyze_all_streams.py <pdf_file>")
        sys.exit(1)
    analyze_all_streams(sys.argv[1])
