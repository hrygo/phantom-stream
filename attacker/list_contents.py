import re
import sys

def list_page_contents(file_path):
    try:
        with open(file_path, 'rb') as f:
            content = f.read()
            content_str = content.decode('latin-1')

        page_pattern = re.compile(r'(\d+)\s+0\s+obj.*?/Type\s*/Page', re.DOTALL)
        
        content_counts = {}

        for match in page_pattern.finditer(content_str):
            obj_id = match.group(1)
            obj_content_pattern = re.compile(r'{}\s+0\s+obj(.*?)endobj'.format(obj_id), re.DOTALL)
            obj_match = obj_content_pattern.search(content_str)
            if not obj_match: continue
            
            obj_data = obj_match.group(1)
            
            # Find Contents
            contents = []
            contents_match = re.search(r'/Contents\s+(\d+)\s+0\s+R', obj_data)
            if contents_match:
                contents.append(contents_match.group(1))
            
            contents_array_match = re.search(r'/Contents\s*\[(.*?)(\s*)\]', obj_data)
            if contents_array_match:
                ids = re.findall(r'(\d+)\s+0\s+R', contents_array_match.group(1))
                contents.extend(ids)
            
            print(f"Page Object {obj_id} Contents: {contents}")
            for cid in contents:
                content_counts[cid] = content_counts.get(cid, 0) + 1

        print("\n--- Content Object Usage Counts ---")
        for cid, count in content_counts.items():
            if count > 1:
                print(f"Object {cid}: Used {count} times")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    list_page_contents(sys.argv[1])
