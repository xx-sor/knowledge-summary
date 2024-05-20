#!/bin/bash

# 检查是否传入了目标文件夹路径
if [ -z "$1" ]; then
  echo "Usage: $0 <relative-path-to-folder>"
  exit 1
fi

# 获取传入的目标文件夹路径
TARGET_FOLDER="$1"

# 检查目标文件夹是否存在
if [ ! -d "$TARGET_FOLDER" ]; then
  echo "Error: Directory $TARGET_FOLDER does not exist."
  exit 1
fi

# 递归遍历目标文件夹及其子文件夹中的所有 Markdown 文件
find "$TARGET_FOLDER" -type f -name "*.md" | while read -r file; do
  echo "Processing $file"
  
  # 在文件中插入目录占位符，如果文件中没有目录占位符
  if ! grep -q "<!-- toc -->" "$file"; then
    echo "Inserting TOC placeholder in $file"
    # 在文件开头插入目录占位符
    sed -i '' '1i\
    <!-- toc -->' "$file"
  fi
  
  # 生成目录
  markdown-toc -i "$file"
done

echo "All files processed."
