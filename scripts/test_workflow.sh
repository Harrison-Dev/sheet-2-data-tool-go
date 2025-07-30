#!/bin/bash

# Excel Schema Generator Test Workflow
# This script demonstrates the complete workflow for processing RPG test data

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Directories
TEST_DATA_DIR="test-data"
TEST_OUTPUT_DIR="test-output"
TEST_SCHEMA_DIR="test-schemas"

echo -e "${BLUE}Excel Schema Generator - Test Workflow${NC}"
echo "======================================="
echo ""

# Clean up previous test outputs
echo -e "${YELLOW}Cleaning up previous test outputs...${NC}"
rm -rf "$TEST_OUTPUT_DIR" "$TEST_SCHEMA_DIR"
mkdir -p "$TEST_OUTPUT_DIR" "$TEST_SCHEMA_DIR"

# Build the project if not already built
if [ ! -f "./data-generator" ]; then
    echo -e "${YELLOW}Building the project...${NC}"
    go build -o data-generator .
fi

# Step 1: Generate initial schema from Excel files
echo ""
echo -e "${GREEN}Step 1: Generating schema from Excel files${NC}"
echo "Command: ./data-generator generate -folder $TEST_DATA_DIR"
./data-generator generate -folder "$TEST_DATA_DIR"

# Move the generated schema to test directory
if [ -f "schema.yml" ]; then
    mv schema.yml "$TEST_SCHEMA_DIR/"
    echo -e "${BLUE}✓ Schema generated and moved to $TEST_SCHEMA_DIR/schema.yml${NC}"
else
    echo -e "${RED}✗ Schema generation failed${NC}"
    exit 1
fi

# Step 2: View the generated schema
echo ""
echo -e "${GREEN}Step 2: Viewing generated schema (first 50 lines)${NC}"
echo "----------------------------------------"
head -n 50 "$TEST_SCHEMA_DIR/schema.yml"
echo "..."
echo "----------------------------------------"

# Step 3: Generate JSON data from schema
echo ""
echo -e "${GREEN}Step 3: Generating JSON data from schema${NC}"
echo "Command: ./data-generator data -folder $TEST_DATA_DIR"
# Copy schema back for data generation
cp "$TEST_SCHEMA_DIR/schema.yml" .
./data-generator data -folder "$TEST_DATA_DIR"

# Move the generated output to test directory
if [ -f "output.json" ]; then
    mv output.json "$TEST_OUTPUT_DIR/"
    echo -e "${BLUE}✓ JSON data generated and moved to $TEST_OUTPUT_DIR/output.json${NC}"
else
    echo -e "${RED}✗ Data generation failed${NC}"
    exit 1
fi

# Step 4: View the generated JSON (first 100 lines)
echo ""
echo -e "${GREEN}Step 4: Viewing generated JSON data (first 100 lines)${NC}"
echo "----------------------------------------"
head -n 100 "$TEST_OUTPUT_DIR/output.json" | jq '.' 2>/dev/null || head -n 100 "$TEST_OUTPUT_DIR/output.json"
echo "..."
echo "----------------------------------------"

# Step 5: Show file sizes and statistics
echo ""
echo -e "${GREEN}Step 5: Test Results Summary${NC}"
echo "----------------------------------------"
echo "Excel files processed:"
ls -lh "$TEST_DATA_DIR"/*.xlsx | awk '{print "  " $9 ": " $5}'
echo ""
echo "Generated files:"
echo "  Schema: $(wc -l < "$TEST_SCHEMA_DIR/schema.yml") lines"
echo "  JSON output: $(wc -l < "$TEST_OUTPUT_DIR/output.json") lines"
echo ""

# Optional: Test update schema functionality
echo -e "${YELLOW}Optional: Test schema update (press Enter to skip, or 'y' to test)${NC}"
read -r response
if [ "$response" = "y" ]; then
    echo ""
    echo -e "${GREEN}Testing schema update functionality${NC}"
    echo "Command: ./data-generator update -folder $TEST_DATA_DIR"
    cp "$TEST_SCHEMA_DIR/schema.yml" .
    ./data-generator update -folder "$TEST_DATA_DIR"
    mv schema.yml "$TEST_SCHEMA_DIR/schema_updated.yml"
    echo -e "${BLUE}✓ Schema update completed${NC}"
fi

# Cleanup
rm -f schema.yml output.json

echo ""
echo -e "${GREEN}Test workflow completed successfully!${NC}"
echo ""
echo "Generated files are located in:"
echo "  - Schema: $TEST_SCHEMA_DIR/"
echo "  - JSON output: $TEST_OUTPUT_DIR/"
echo ""
echo "To run the GUI version, simply execute:"
echo "  ./data-generator"