package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func mergeSwagger(docs ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, doc := range docs {
		for k, v := range doc {
			if existing, ok := result[k]; ok {
				switch k {
				case "paths", "definitions":
					// merge map[string]interface{}
					if em, ok := existing.(map[string]interface{}); ok {
						if vm, ok := v.(map[string]interface{}); ok {
							for kk, vv := range vm {
								em[kk] = vv
							}
						}
						// 排序 map
						result[k] = sortMap(em)
					}
				case "tags":
					// 去重後合併 tags
					if es, ok := existing.([]interface{}); ok {
						if vs, ok := v.([]interface{}); ok {
							merged := dedupTags(append(es, vs...))
							result[k] = sortTags(merged)
						}
					}
				default:
					// overwrite
					result[k] = v
				}
			} else {
				switch k {
				case "paths", "definitions":
					if vm, ok := v.(map[string]interface{}); ok {
						result[k] = sortMap(vm)
						continue
					}
				case "tags":
					if vs, ok := v.([]interface{}); ok {
						result[k] = sortTags(vs)
						continue
					}
				}
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

// sortMap 把 map[string]interface{} 的 key 排序
func sortMap(m map[string]interface{}) map[string]interface{} {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	newMap := make(map[string]interface{}, len(m))
	for _, k := range keys {
		newMap[k] = m[k]
	}
	return newMap
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
					return fmt.Errorf("failed to read file %s: %w", f, err)
				}
				var m map[string]interface{}
				if err := yaml.Unmarshal(data, &m); err != nil {
					return fmt.Errorf("failed to parse yaml %s: %w", f, err)
				}
				docs = append(docs, m)
			}

			merged := mergeSwagger(docs...)

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
