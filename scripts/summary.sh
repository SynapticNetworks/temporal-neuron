#!/bin/bash

# Define output file
OUTPUT_FILE="summary.txt"

# Check for .summary.ignore file
if [ ! -f ".summary.ignore" ]; then
  echo "No .summary.ignore file found. Exiting."
  exit 1
fi

# Generate the tree output with exclusions
echo "Generating directory tree..."
# Create a pattern for tree exclusion
TREE_EXCLUDE="go.mod|go.sum|.gitignore|.git|node_modules|.DS_Store|out|build|dist|.plasmo|coverage|*.env|*.pem"

# Add patterns from .summary.ignore to tree exclude
while IFS= read -r line || [ -n "$line" ]; do
  # Skip comments and empty lines
  if [[ ! "$line" =~ ^[[:space:]]*# && -n "${line// }" ]]; then
    # Remove leading slash if present
    clean_line="${line#/}"
    # Remove trailing slash if present
    clean_line="${clean_line%/}"
    
    # Only add non-empty patterns
    if [ -n "$clean_line" ]; then
      TREE_EXCLUDE="$TREE_EXCLUDE|$clean_line"
    fi
  fi
done < .summary.ignore

echo "Tree exclude pattern: $TREE_EXCLUDE"
tree -a -I "$TREE_EXCLUDE" > "$OUTPUT_FILE"

# Build the find command with exclusions
FIND_CMD="find . -type f"
FIND_CMD="$FIND_CMD ! -name go.mod ! -name go.sum ! -name .gitignore ! -name summary.txt"
FIND_CMD="$FIND_CMD ! -path \"./.git/*\" ! -path \"./node_modules/*\""

# Add explicit exclusions for common patterns from your ignore file
FIND_CMD="$FIND_CMD ! -path \"*/.DS_Store\" ! -path \"*/out/*\" ! -path \"*/build/*\" ! -path \"*/dist/*\""
FIND_CMD="$FIND_CMD ! -path \"*/.plasmo/*\" ! -path \"*/coverage/*\""
FIND_CMD="$FIND_CMD ! -path \"*/*.env\""

# Add each pattern from .summary.ignore to the find command
while IFS= read -r line || [ -n "$line" ]; do
  # Skip comments and empty lines
  if [[ ! "$line" =~ ^[[:space:]]*# && -n "${line// }" ]]; then
    # Remove leading slash if present (e.g., /node_modules -> node_modules)
    clean_line="${line#/}"
    
    # Handle paths differently than file patterns
    if [[ "$clean_line" == */* ]]; then
      # For directory patterns with trailing slash, exclude everything in that directory
      if [[ "$line" == */ ]]; then
        FIND_CMD="$FIND_CMD ! -path \"*/${clean_line}*\""
      else
        FIND_CMD="$FIND_CMD ! -path \"*/${clean_line}\""
      fi
    else
      # For patterns that look like directories without slashes (e.g. .plasmo)
      if [[ "$clean_line" == .* || -d "$clean_line" ]]; then
        FIND_CMD="$FIND_CMD ! -path \"*/$clean_line/*\" ! -path \"*/$clean_line\""
      # For wildcard patterns (e.g. *.env, *.pem)
      elif [[ "$clean_line" == *\** ]]; then
        pattern_part="${clean_line//\*/.*}"
        FIND_CMD="$FIND_CMD ! -regex \".*${pattern_part}\""
      else
        FIND_CMD="$FIND_CMD ! -name \"$clean_line\""
      fi
    fi
  fi
done < .summary.ignore

# Debug - remove this line when confirmed working
echo "Find command: $FIND_CMD"

# Process files
echo "Processing files..."
eval "$FIND_CMD" | while read -r file; do
  # Skip directories that might still be included
  if [ -d "$file" ]; then
    continue
  fi
  
  echo "Adding: $file"
  
  # Append file path as a header
  echo "#### $file ####" >> "$OUTPUT_FILE"
  
  # Append file content
  cat "$file" >> "$OUTPUT_FILE"
  
  # Add a newline for separation
  echo "" >> "$OUTPUT_FILE"
done

echo "Done! Output saved to $OUTPUT_FILE"