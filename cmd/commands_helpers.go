package main

import (
	schemaDomain "github.com/lexlapax/go-llms/pkg/schema/domain"
)

// Helper functions for schema conversion

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getProperties(m map[string]interface{}) map[string]schemaDomain.Property {
	props := make(map[string]schemaDomain.Property)
	
	if propMap, ok := m["properties"].(map[string]interface{}); ok {
		for key, val := range propMap {
			if propSchema, ok := val.(map[string]interface{}); ok {
				props[key] = schemaDomain.Property{
					Type:        getString(propSchema, "type"),
					Description: getString(propSchema, "description"),
					Properties:  getProperties(propSchema),
					Required:    getRequired(propSchema),
					Items:       getItemsProperty(propSchema),
				}
			}
		}
	}
	
	return props
}

func getRequired(m map[string]interface{}) []string {
	var required []string
	
	if reqArray, ok := m["required"].([]interface{}); ok {
		for _, item := range reqArray {
			if str, ok := item.(string); ok {
				required = append(required, str)
			}
		}
	}
	
	return required
}

func getItemsProperty(m map[string]interface{}) *schemaDomain.Property {
	if itemsMap, ok := m["items"].(map[string]interface{}); ok {
		return &schemaDomain.Property{
			Type:       getString(itemsMap, "type"),
			Properties: getProperties(itemsMap),
			Required:   getRequired(itemsMap),
			Items:      getItemsProperty(itemsMap),
		}
	}
	return nil
}