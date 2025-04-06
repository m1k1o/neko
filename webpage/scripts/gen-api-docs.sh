#!/bin/bash
cd "$(dirname "$0")/.."

# Clean the API docs
docusaurus clean-api-docs all

# Generate the API docs
docusaurus gen-api-docs all

# Create README.md
mv docs/api/neko-api.info.mdx docs/api/README.mdx

# Replace all occurences of docs/api/neko-api with docs/v3/api
find docs/api -type f -exec sed -i 's/docs\/api\/neko-api/docs\/v3\/api/g' {} \;

# This regex removes (multiline) any span tag that contains "theme-doc-version-badge":
sed -i '/<span/{:a;N;/<\/span>/!ba;/theme-doc-version-badge/d}' docs/api/README.mdx
