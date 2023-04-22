#!/bin/bash

rm -rf dist/types

# Find all vue files, and convert them to .ts files
find src/component -name "*.vue" -type f -print0 | while IFS= read -r -d '' file; do
    # Get the file name
    filename=$(basename -- "$file")
    # Get the file name without extension
    filename="${filename%.*}"
    # Get the directory name
    directory=$(dirname -- "$file")
    # Get the directory name without the first dot
    directory="${directory}"

    if [ "$directory" = "" ]; then
      directory="."
    fi

    echo "Creating: $file --> $directory/$filename.ts"

    # Get contant of the file, that is between <script lang="ts"> and </script>
    content=$(sed -n '/<script lang="ts">/,/<\/script>/p' "$file" | sed '1d;$d')

    # Remove .vue in imports
    content=$(echo "$content" | sed 's/\.vue//g')

    # Create a new file with the same name and .ts extension
    echo "$content" > "$directory/$filename.ts"

    # Add file to .toDelete file
    echo "$directory/$filename.ts" >> .toDelete
done

echo "Compiling files"
npx tsc -p types-tsconfig.json

# Remove all .js files in dist/types folder
echo "Removing .js files in dist/types"
find dist/types -name "*.js" -type f -delete

# Remove all files listed in .toDelete
echo "Removing files listed in .toDelete"
while read -r file; do
  echo "Removing: $file"
  rm -f "$file"
done < .toDelete

# Remove .toDelete file
rm -f .toDelete
