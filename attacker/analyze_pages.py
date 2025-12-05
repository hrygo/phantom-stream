import re
import zlib
import sys

def analyze_pages(file_path):
    try:
        with open(file_path, 'rb') as f:
            content = f.read()
            content_str = content.decode('latin-1')

        # 1. Find all Page objects
        # Look for /Type /Page
        page_pattern = re.compile(r'(\d+)\s+0\s+obj.*?/Type\s*/Page', re.DOTALL)
        
        for match in page_pattern.finditer(content_str):
            obj_id = match.group(1)
            print(f"[*] Analyzing Page Object {obj_id}...")
            
            # Find the object content
            obj_content_pattern = re.compile(r'{}\s+0\s+obj(.*?)endobj'.format(obj_id), re.DOTALL)
            obj_match = obj_content_pattern.search(content_str)
            if not obj_match:
                continue
            
            obj_data = obj_match.group(1)
            
            # Find Contents
            contents_match = re.search(r'/Contents\s+(\d+)\s+0\s+R', obj_data)
            if contents_match:
                contents_id = contents_match.group(1)
                print(f"    -> Contents Object: {contents_id}")
                analyze_content_stream(content_str, contents_id)
            
            # Handle array of contents if any (rare in simple PDFs but possible)
            contents_array_match = re.search(r'/Contents\s*\[(.*?)\]', obj_data)
            if contents_array_match:
                ids = re.findall(r'(\d+)\s+0\s+R', contents_array_match.group(1))
                print(f"    -> Contents Objects: {ids}")
                for cid in ids:
                    analyze_content_stream(content_str, cid)

    except Exception as e:
        print(f"Error: {e}")

def analyze_content_stream(pdf_content, obj_id):
    stream_pattern = re.compile(r'{}\s+0\s+obj[\s\S]*?>>\s*stream\s*([\s\S]*?)\s*endstream'.format(obj_id), re.DOTALL)
    match = stream_pattern.search(pdf_content)
    if not match:
        print(f"    [!] Stream for object {obj_id} not found")
        return

    stream_data = match.group(1).encode('latin-1')
    try:
        # Try decompressing
        decompressed = zlib.decompress(stream_data)
        # Search for watermark text
        text = decompressed.decode('latin-1', errors='ignore')
        if "CONFIDENTIAL" in text or "CAFEBABE" in text:
            print(f"    [!!!] FOUND WATERMARK IN OBJECT {obj_id}")
            # Print context
            idx = text.find("CONFIDENTIAL")
            if idx == -1: idx = text.find("CAFEBABE")
            start = max(0, idx - 50)
            end = min(len(text), idx + 100)
            print(f"    Context: ...{text[start:end]}...")
        else:
            print(f"    [+] Object {obj_id} stream clean (no plain text watermark).")
            
    except Exception as e:
        print(f"    [!] Decompression failed for object {obj_id}: {e}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 analyze_pages.py <pdf_file>")
        sys.exit(1)
    analyze_pages(sys.argv[1])
