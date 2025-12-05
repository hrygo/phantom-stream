import re
import sys

def find_smask_objects(file_path):
    try:
        with open(file_path, 'rb') as f:
            content = f.read()
            content_str = content.decode('latin-1')  # PDF uses binary, but latin-1 preserves bytes

        print(f"Analyzing {file_path}...")

        # Find all objects with /SMask
        # Pattern: look for object definitions that contain /SMask
        # This is a simple regex, might need refinement for complex PDFs
        
        # 1. Find all "obj ... endobj" blocks
        obj_pattern = re.compile(r'(\d+)\s+0\s+obj(.*?)endobj', re.DOTALL)
        
        objects = []
        for match in obj_pattern.finditer(content_str):
            obj_id = match.group(1)
            obj_content = match.group(2)
            
            if '/SMask' in obj_content:
                print(f"[!] Found Object {obj_id} with /SMask")
                
                # Extract the SMask reference if it's a reference
                smask_ref_match = re.search(r'/SMask\s+(\d+)\s+0\s+R', obj_content)
                if smask_ref_match:
                    smask_id = smask_ref_match.group(1)
                    print(f"    -> References SMask Object: {smask_id}")
                    objects.append((obj_id, smask_id))
                else:
                    print(f"    -> SMask might be inline or not a direct reference")

        return objects

    except Exception as e:
        print(f"Error: {e}")
        return []

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 find_smask.py <pdf_file>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    find_smask_objects(file_path)
