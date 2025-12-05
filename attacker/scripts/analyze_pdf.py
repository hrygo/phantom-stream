import zlib
import re

def calculate_entropy(data):
    """Calculate Shannon entropy of data"""
    if not data:
        return 0
    
    freq = {}
    for byte in data:
        freq[byte] = freq.get(byte, 0) + 1
    
    entropy = 0.0
    data_len = len(data)
    for count in freq.values():
        p = count / data_len
        if p > 0:
            entropy -= p * math.log2(p)
    
    return entropy

# Read the file
with open('test_data/2511.17467v2_Phase7_R2_signed.pdf', 'rb') as f:
    pdf_data = f.read()

# Find all image objects
content = pdf_data.decode('latin-1')
import math

# Find object 32 specifically
obj32_start = content.find('32 0 obj')
if obj32_start != -1:
    obj32_content = content[obj32_start:obj32_start+500]
    
    # Find stream
    stream_match = re.search(r'stream\s*(.*?)\s*endstream', obj32_content, re.DOTALL)
    if stream_match:
        stream_data = stream_match.group(1).encode('latin-1')
        
        # Check if it's compressed
        if stream_data.startswith(b'x\x9c'):
            try:
                # Decompress
                decompressed = zlib.decompress(stream_data[2:-4])
                print(f'Object 32:')
                print(f'  Compressed size: {len(stream_data)}')
                print(f'  Decompressed size: {len(decompressed)}')
                
                # Sample entropy
                sample = decompressed[:1000] if len(decompressed) > 1000 else decompressed
                entropy = calculate_entropy(sample)
                print(f'  Sample entropy: {entropy:.2f}')
                
                # Check for suspicious patterns
                if len(decompressed) >= 10:
                    # Look for any non-image-like patterns
                    unique_bytes = len(set(decompressed[:100]))
                    print(f'  Unique bytes in first 100: {unique_bytes}')
                    
                    # Look for potential signatures
                    for i in range(len(decompressed)-10):
                        chunk = decompressed[i:i+11]
                        # Check for printable ASCII patterns
                        if all(32 <= b <= 126 for b in chunk):
                            print(f'  Suspicious ASCII at offset {i}: {chunk}')
                            
            except Exception as e:
                print(f'Error decompressing object 32: {e}')
