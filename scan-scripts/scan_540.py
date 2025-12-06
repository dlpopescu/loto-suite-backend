import pytesseract
import re
from PIL import Image
import sys
import json

def extract_540_info(image_path, debug=False):
    """Extract lottery info specifically for SUPER LOTO 5/40 tickets"""
    # OCR the image
    text = pytesseract.image_to_string(Image.open(image_path), lang='eng')
    lines = [line.strip().upper() for line in text.splitlines() if line.strip()]
    
    if debug:
        print("\n=== RAW OCR OUTPUT ===")
        for i, line in enumerate(lines, 1):
            print(f"{i}: {repr(line)}")
        print("======================\n")
    
    result = {'game': '5/40'}
    
    # Extract draw date: DT:DD/MM/YY
    result['draw_date'] = None
    for line in lines:
        m = re.search(r'DT[:\s;]*(\d{2}[/|]\d{2}[/|]\d{2})', line)
        if m:
            result['draw_date'] = m.group(1).replace('|', '/')
            break
    
    # Extract variants: A05, B05, C05, D05 (max 4 variants)
    variants = []
    seen_ids = set()
    
    # Pass 1: Try to match with letter identifiers
    for line in lines:
        # First try: Match "A05: 01,10,12,26,29" pattern
        # Handle OCR errors like "PAO5", "BBO5", "BCO5" by allowing junk before the letter
        m = re.search(r'[^A-D]*([A-D])(?:O|0)(?:5|S)[:\s;]+([0-9\s,]+)', line)
        if m and m.group(1) not in seen_ids:
            nums = [int(n) for n in re.findall(r'\d+', m.group(2))]
            if len(nums) == 5 and all(1 <= n <= 40 for n in nums):
                variants.append({'id': m.group(1), 'numbers': nums})
                seen_ids.add(m.group(1))
                if debug:
                    print(f"Pass 1: Matched variant {m.group(1)}: {nums}")
                continue
        
        # Second try: More flexible pattern allowing noise between letter and numbers
        m2 = re.search(r'[^A-D]*([A-D])[^A-D:\d;]{0,6}[:\s;]+([0-9\s,]+)', line)
        if m2 and m2.group(1) not in seen_ids:
            nums = [int(n) for n in re.findall(r'\d+', m2.group(2))]
            if len(nums) == 5 and all(1 <= n <= 40 for n in nums):
                variants.append({'id': m2.group(1), 'numbers': nums})
                seen_ids.add(m2.group(1))
                if debug:
                    print(f"Pass 2: Matched variant {m2.group(1)}: {nums}")
                continue
    
    # Pass 2: Fallback - look for any lines with exactly 5 numbers in valid range
    # This catches variants where the letter was misread or missing
    if len(variants) < 4:
        if debug:
            print(f"Pass 3: Only found {len(variants)} variants, looking for more...")
        
        for line in lines:
            if len(variants) >= 4:
                break
            
            # Skip lines we've already matched
            already_matched = False
            for v in variants:
                nums_str = ','.join(str(n) for n in v['numbers'])
                if nums_str in line.replace(' ', ''):
                    already_matched = True
                    break
            if already_matched:
                continue
            
            # Look for exactly 5 numbers in sequence
            nums = [int(n) for n in re.findall(r'\d+', line)]
            if len(nums) == 5 and all(1 <= n <= 40 for n in nums):
                # Check this isn't a duplicate of an existing variant
                is_duplicate = False
                for v in variants:
                    if v['numbers'] == nums:
                        is_duplicate = True
                        break
                
                if not is_duplicate:
                    # Assign to next missing letter
                    for letter in 'ABCD':
                        if letter not in seen_ids:
                            variants.append({'id': letter, 'numbers': nums})
                            seen_ids.add(letter)
                            if debug:
                                print(f"Pass 3: Fallback matched variant {letter}: {nums}")
                            break
    
    result['variants'] = variants
    
    # Extract Super Noroc: SUPER NOROC: XXXXXX (6 digits)
    result['noroc'] = None
    for line in lines:
        m = re.search(r'SUPER\s+NOROC[^:]*:\s*(\d{6})', line)
        if m:
            result['noroc'] = m.group(1)
            break
    
    return result

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print("Usage: python ticket_540.py <image_path> [--debug] [--json]")
        print("\nExtract lottery info from SUPER LOTO 5/40 tickets")
        print("  --debug: Show raw OCR output")
        print("  --json: Output in JSON format (default: human-readable)")
        sys.exit(1)
    
    image_path = sys.argv[1]
    debug = '--debug' in sys.argv
    json_output = '--json' in sys.argv
    
    info = extract_540_info(image_path, debug=debug)
    
    if json_output:
        # Convert to the cache format
        # Map variant letters (A, B, C, D) to 1-based numeric indices (1, 2, 3, 4)
        # Frontend expects 1-based IDs to match input element IDs like number-input-1-1
        letter_to_index = {'A': 1, 'B': 2, 'C': 3, 'D': 4}
        output = {
            "game_id": "540",
            "game_date": info.get('draw_date', ''),
            "nume_noroc": "SUPER NOROC",
            "variante": [
                {
                    "id": letter_to_index.get(v['id'], 1),
                    "numere": [
                        {
                            "numar": str(num),
                        }
                        for num in v['numbers']
                    ]
                }
                for v in info.get('variants', [])
            ],
            "noroc": {
                "numar": info.get('noroc', '')
            }
        }
        print(json.dumps(output, indent=2, ensure_ascii=False))
    else:
        # Human-readable format
        print("\n" + "="*50)
        print("SUPER LOTO 5/40 TICKET SCAN RESULTS")
        print("="*50)
        print(f"Draw Date: {info.get('draw_date', 'N/A')}")
        print(f"\nVariants ({len(info.get('variants', []))} detected):")
        for v in info.get('variants', []):
            nums_str = ', '.join(map(str, v['numbers']))
            print(f"  {v['id']}: [{nums_str}]")
        print(f"\nSuper Noroc (6 digits): {info.get('noroc', 'N/A')}")
        print("="*50 + "\n")
