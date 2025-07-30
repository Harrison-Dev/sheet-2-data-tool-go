#!/bin/bash

# Complete Test Script for Excel Schema Generator
set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}Excel Schema Generator - Complete Test${NC}"
echo "======================================="
echo ""

# Build if needed
if [ ! -f "./data-generator" ]; then
    echo -e "${YELLOW}Building project...${NC}"
    go build -o data-generator .
fi

# Clean up
rm -rf test-output test-schemas
mkdir -p test-output test-schemas
rm -f schema.yml output.json

echo -e "${GREEN}Step 1: Generate initial schema${NC}"
./data-generator generate -folder test-data

echo -e "${GREEN}Step 2: Fix offset_header (change from 2 to 1)${NC}"
sed -i '' 's/offset_header: 2/offset_header: 1/g' schema.yml

echo -e "${GREEN}Step 3: Update schema with field information${NC}"
./data-generator update -folder test-data

echo -e "${GREEN}Step 4: Fix Id field type (change from string to int)${NC}"
sed -i '' 's/name: Id$/&\
          data_type: int/g; s/name: Id\
          data_type: string/name: Id\
          data_type: int/g' schema.yml

echo -e "${GREEN}Step 5: Generate JSON data${NC}"
./data-generator data -folder test-data

# Move files to test directories
mv schema.yml test-schemas/
mv output.json test-output/

echo ""
echo -e "${GREEN}✅ Test completed successfully!${NC}"
echo ""
echo "Generated files:"
echo "  📄 Schema: test-schemas/schema.yml ($(wc -l < test-schemas/schema.yml) lines)"
echo "  📊 JSON: test-output/output.json ($(wc -l < test-output/output.json) lines)"
echo ""

echo -e "${BLUE}Viewing schema structure:${NC}"
echo "----------------------------------------"
head -n 30 test-schemas/schema.yml
echo "..."
echo "----------------------------------------"
echo ""

echo -e "${BLUE}Viewing JSON output sample:${NC}"
echo "----------------------------------------"
if command -v jq &> /dev/null; then
    head -n 50 test-output/output.json | jq '.' | head -n 30
else
    head -n 30 test-output/output.json
fi
echo "..."
echo "----------------------------------------"
echo ""

echo -e "${GREEN}Test Summary:${NC}"
echo "✓ Schema generation: Success"
echo "✓ Schema update: Success"  
echo "✓ Data generation: Success"
echo ""
echo "Test data files:"
ls -la test-data/*.xlsx | awk '{print "  " $9 ": " $5}'
echo ""
echo "To view full output:"
echo "  cat test-schemas/schema.yml"
echo "  cat test-output/output.json | jq '.'"