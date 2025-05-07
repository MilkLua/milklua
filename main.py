#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os

def main():
    output_file = "output.txt"
    # 打开输出文件(以写模式，使用utf-8编码)
    with open(output_file, "w", encoding="utf-8") as out_f:
        # 使用 os.walk 递归遍历当前目录
        for root, dirs, files in os.walk("."):
            for filename in files:
                # 检查文件扩展名是否为 .go
                if filename.endswith(".go"):
                    # 生成相对路径，为了使格式与示例一致，可加上"./"前缀
                    rel_path = os.path.join(root, filename)
                    # 如果 rel_path 不以"./"开始，则添加
                    if not rel_path.startswith("."):
                        rel_path = "./" + rel_path
                    
                    # 写入文件的相对路径
                    out_f.write(rel_path + "\n")
                    
                    # 读取 .go 文件内容，并写入到输出文件
                    try:
                        with open(rel_path, "r", encoding="utf-8") as in_f:
                            content = in_f.read()
                            out_f.write(content)
                    except Exception as e:
                        out_f.write(f"读取文件时出错: {e}")
                    
                    # 写入四个空行作为分隔
                    out_f.write("\n" * 4)

if __name__ == '__main__':
    main()