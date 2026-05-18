import pandas as pd
import json
import uuid
import os

# Models to process
models = [
    {
        "model_code": "82SN003JVN",
        "as_built": "../data/IdeaPad 5 Pro 16ARH7 - Type 82SN_AsBuilt.xlsx",
        "model": "../data/IdeaPad 5 Pro 16ARH7 - Type 82SN_Model.xlsx"
    },
    {
        "model_code": "83LY00HQVN",
        "as_built": "../data/Legion 5 15IRX10 - Type 83LY_AsBuilt.xlsx",
        "model": "../data/Legion 5 15IRX10 - Type 83LY_Model.xlsx"
    }
]

# We need:
# 1. PartTypes: list of dicts { "commodity_type": str, "description": str }
# 2. PartCatalogs: list of dicts { "part_number": str, "commodity_type": str, "mfg_number": str, "description": str }
# 3. Installations: dict mapping `model_code` -> list of dicts { "part_number": str, "quantity": int, "mfg_number": str }

part_types_set = set() # commodity_type
part_catalogs_dict = {} # (part_number, commodity_type, mfg_number) -> description

installations = {}

for m in models:
    model_code = m["model_code"]
    as_built_df = pd.read_excel(m["as_built"])
    model_df = pd.read_excel(m["model"])

    installations[model_code] = []

    # Process Model.xlsx for PartTypes
    for _, row in model_df.iterrows():
        ct = str(row.get('Commodity Type', '')).strip()
        if ct and ct != 'nan':
            part_types_set.add(ct)

    # Process AsBuilt.xlsx for PartCatalog & Installations
    for _, row in as_built_df.iterrows():
        pn = str(row.get('Part Number', '')).strip()
        ct = str(row.get('Commodity Type', '')).strip()
        desc = str(row.get('Description', '')).strip()
        qty = row.get('Installed Qty', 1)

        try:
            qty = int(qty)
        except:
            qty = 1

        mfgs = str(row.get('MFG Part Number', '')).strip()

        if pn and pn != 'nan' and ct and ct != 'nan':
            part_types_set.add(ct)

            # Split MFGs by comma
            mfg_list = [x.strip() for x in mfgs.split(',')] if mfgs and mfgs != 'nan' else ['UNKNOWN']

            for mfg in mfg_list:
                key = (pn, ct, mfg)
                # Take description from AsBuilt, if multiple rows for same key, last one wins
                if desc and desc != 'nan':
                    part_catalogs_dict[key] = desc
                elif key not in part_catalogs_dict:
                    part_catalogs_dict[key] = "Generic " + ct

                installations[model_code].append({
                    "part_number": pn,
                    "mfg_number": mfg,
                    "quantity": qty
                })

# Convert to list
out_part_types = [{"commodity_type": ct, "description": ""} for ct in sorted(list(part_types_set))]
out_part_catalogs = [
    {
        "part_number": pn, 
        "commodity_type": ct, 
        "mfg_number": mfg, 
        "description": desc
    } for (pn, ct, mfg), desc in part_catalogs_dict.items()
]

out_data = {
    "part_types": out_part_types,
    "part_catalogs": out_part_catalogs,
    "installations": installations
}

os.makedirs("resources", exist_ok=True)
with open("resources/parts.json", "w", encoding="utf-8") as f:
    json.dump(out_data, f, indent=4)

print(f"Generated resources/parts.json with {len(out_part_types)} PartTypes, {len(out_part_catalogs)} PartCatalogs, and mapped to {len(installations)} Models.")
