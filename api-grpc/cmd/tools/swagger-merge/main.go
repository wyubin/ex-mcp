package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/wyubin/ex-mcp/api-grpc/src/utils/maptool"
	"gopkg.in/yaml.v3"
)

var orderKey = []string{"swagger", "info", "host", "basePath", "schemes", "consumes", "produces", "tags", "paths", "definitions"}

func mergeSwagger(docs ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, doc := range docs {
		for k, v := range doc {
			existData, exists := result[k]
			switch k {
			case "paths", "definitions":
				em := maptool.NewOrderedMap[any]()
				if exists { // 已存在前一個 yaml 資料
					em = existData.(maptool.OrderedMap[any])
				}
				if vm, ok := v.(map[string]interface{}); ok { // 把每個 item 塞入
					for kk, vv := range vm {
						em.Append(kk, vv)
					}
				}
				em.OrderBy()
				result[k] = em
			case "tags":
				es := []interface{}{}
				if exists { // 已存在前一個 yaml 資料
					es = existData.([]interface{})
				}
				// 去重後合併 tags
				if vs, ok := v.([]interface{}); ok {
					result[k] = sortTags(dedupTags(append(es, vs...)))
				}
			default:
				// overwrite
				result[k] = v
			}
		}
	}
	return result
}

// dedupTags 依照 tag 的 name 去重
func dedupTags(tags []interface{}) []interface{} {
	seen := make(map[string]bool)
	var result []interface{}

	for _, t := range tags {
		if m, ok := t.(map[string]interface{}); ok {
			if name, ok := m["name"].(string); ok {
				if seen[name] {
					continue
				}
				seen[name] = true
				result = append(result, m)
			}
		} else {
			// 非預期型態就直接塞
			result = append(result, t)
		}
	}
	return result
}

// sortTags 依照 name 排序
func sortTags(tags []interface{}) []interface{} {
	sort.Slice(tags, func(i, j int) bool {
		mi, iok := tags[i].(map[string]interface{})
		mj, jok := tags[j].(map[string]interface{})
		if iok && jok {
			ni, _ := mi["name"].(string)
			nj, _ := mj["name"].(string)
			return ni < nj
		}
		return false
	})
	return tags
}

func main() {
	var output string

	rootCmd := &cobra.Command{
		Use:   "swagger-tool",
		Short: "Swagger YAML merge tool",
	}

	mergeCmd := &cobra.Command{
		Use:   "merge [files...]",
		Short: "Merge multiple swagger yaml files",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var docs []map[string]interface{}
			for _, f := range args {
				data, err := os.ReadFile(f)
				if err != nil {
					fmt.Fprintf(os.Stderr, "failed to read file %s: %s", f, err)
					continue
				}
				var m map[string]interface{}
				if err := yaml.Unmarshal(data, &m); err != nil {
					fmt.Fprintf(os.Stderr, "failed to parse yaml %s: %s", f, err)
					continue
				}
				docs = append(docs, m)
			}
			merged := maptool.NewOrderedMap[any]()
			for k, v := range mergeSwagger(docs...) {
				merged.Append(k, v)
			}
			merged.OrderBy(orderKey...)

			out, err := yaml.Marshal(merged)
			if err != nil {
				return fmt.Errorf("failed to marshal merged yaml: %w", err)
			}

			if output != "" {
				if err := os.WriteFile(output, out, 0644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("Merged swagger written to %s\n", output)
			} else {
				fmt.Println(string(out))
			}
			return nil
		},
	}

	mergeCmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default stdout)")
	rootCmd.AddCommand(mergeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
