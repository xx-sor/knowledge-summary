### 详细注释

1. **检查是否传入了目标文件夹路径**：
    ```sh
    if [ -z "$1" ]; then
      echo "Usage: $0 <relative-path-to-folder>"
      exit 1
    fi
    ```
    - `if [ -z "$1" ]; then`：检查第一个参数是否为空。
    - `echo "Usage: $0 <relative-path-to-folder>"`：如果参数为空，输出使用说明。
    - `exit 1`：退出脚本，返回状态码 1。

2. **获取传入的目标文件夹路径**：
    ```sh
    TARGET_FOLDER="$1"
    ```
    - `TARGET_FOLDER="$1"`：将第一个参数赋值给 `TARGET_FOLDER` 变量。

3. **检查目标文件夹是否存在**：
    ```sh
    if [ ! -d "$TARGET_FOLDER" ]; then
      echo "Error: Directory $TARGET_FOLDER does not exist."
      exit 1
    fi
    ```
    - `if [ ! -d "$TARGET_FOLDER" ]; then`：检查 `TARGET_FOLDER` 是否是一个目录。
    - `echo "Error: Directory $TARGET_FOLDER does not exist."`：如果不是目录，输出错误信息。
    - `exit 1`：退出脚本，返回状态码 1。

4. **递归遍历目标文件夹及其子文件夹中的所有 Markdown 文件**：
    ```sh
    find "$TARGET_FOLDER" -type f -name "*.md" | while read -r file; do
      echo "Processing $file"
    ```
    - `find "$TARGET_FOLDER" -type f -name "*.md"`：使用 `find` 命令递归查找目标文件夹及其子文件夹中的所有 Markdown 文件。
    - `| while read -r file; do`：将查找到的每个文件传递给 `while` 循环处理。

5. **在文件中插入目录占位符**：
    ```sh
    if ! grep -q "<!-- toc -->" "$file"; then
      echo "Inserting TOC placeholder in $file"
      sed -i '1i <!-- toc -->' "$file"
    fi
    ```
    - `if ! grep -q "<!-- toc -->" "$file"; then`：检查文件中是否已经包含目录占位符 `<!-- toc -->`。
    - `echo "Inserting TOC placeholder in $file"`：如果文件中没有目录占位符，输出插入信息。
    - `sed -i '1i <!-- toc -->' "$file"`：在文件开头插入目录占位符。这是Linux的用法。mac的用法是：
    - `sed -i '' '1i\<!-- toc -->' "$file"`

6. **生成目录**：
    ```sh
    markdown-toc -i "$file"
    ```
    - `markdown-toc -i "$file"`：为文件生成目录。

7. **结束循环和脚本**：
    ```sh
    done

    echo "All files processed."
    ```
    - `done`：结束 `while` 循环。
    - `echo "All files processed."`：输出所有文件处理完成的信息。

### 使用脚本

1. 将上述脚本保存为 `generate_toc.sh`。
2. 赋予脚本执行权限：

    ```sh
    chmod +x generate_toc.sh
    ```

3. 运行脚本，并传入目标文件夹的相对路径：

    ```sh
    ./generate_toc.sh ./path/to/your/markdown/files
    ```

通过这个脚本，你可以递归遍历目标文件夹及其子文件夹中的所有 Markdown 文件，并为每个文件生成目录。
