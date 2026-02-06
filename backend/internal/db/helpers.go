package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/google/uuid"
)

func Insert(ctx context.Context, table string, model interface{}) error {
	db := GetDB()
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer")
	}
	v = v.Elem()

	idField := v.FieldByName("ID")
	if !idField.IsValid() || idField.Type() != reflect.TypeOf(uuid.UUID{}) {
		return fmt.Errorf("model must have UUID ID field")
	}

	if idField.Interface().(uuid.UUID) == uuid.Nil {
		idField.Set(reflect.ValueOf(uuid.New()))
	}

	now := time.Now()
	createdAtField := v.FieldByName("CreatedAt")
	if createdAtField.IsValid() && createdAtField.Type() == reflect.TypeOf(time.Time{}) {
		if createdAtField.Interface().(time.Time).IsZero() {
			createdAtField.Set(reflect.ValueOf(now))
		}
	}

	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}

	idTag := "id"
	columns = append(columns, idTag)
	placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)+1))
	values = append(values, idField.Interface())

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		if field.Name == "ID" || field.Name == "CreatedAt" || field.Name == "UpdatedAt" {
			continue
		}

		columns = append(columns, dbTag)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(values)+1))
		values = append(values, v.Field(i).Interface())
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, join(columns, ", "), join(placeholders, ", "))
	log.Printf("INSERT: %s values=%v", query, values)
	_, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		log.Printf("INSERT error: %v", err)
	}
	return err
}

func Update(ctx context.Context, table string, model interface{}) error {
	db := GetDB()
	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer")
	}
	v = v.Elem()

	idField := v.FieldByName("ID")
	if !idField.IsValid() || idField.Type() != reflect.TypeOf(uuid.UUID{}) {
		return fmt.Errorf("model must have UUID ID field")
	}

	id := idField.Interface().(uuid.UUID)

	sets := []string{}
	values := []interface{}{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		if field.Name == "ID" || field.Name == "CreatedAt" {
			continue
		}

		sets = append(sets, fmt.Sprintf("%s = $%d", dbTag, len(sets)+1))
		values = append(values, v.Field(i).Interface())
	}

	if table == "jobs" || table == "changesets" || table == "review_threads" || table == "review_comments" || table == "chats" || table == "chat_messages" {
		sets = append(sets, fmt.Sprintf("updated_at = $%d", len(sets)+1))
		values = append(values, time.Now())
	}

	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", table, join(sets, ", "), len(values))
	log.Printf("[db.Update] Query: %s", query)
	log.Printf("[db.Update] Values: %+v", values)
	_, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		log.Printf("[db.Update] Error: %v", err)
	} else {
		log.Printf("[db.Update] Success")
	}
	return err
}

func Get(ctx context.Context, model interface{}, query string, args ...interface{}) error {
	db := GetDB()
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return scan(rows, model)
	}
	return sql.ErrNoRows
}

func Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	db := GetDB()
	return db.QueryContext(ctx, query, args...)
}

func Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	db := GetDB()
	return db.ExecContext(ctx, query, args...)
}

func scan(rows *sql.Rows, model interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(model)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("model must be a pointer")
	}
	v = v.Elem()

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return err
	}

	for i, col := range columns {
		for j := 0; j < v.NumField(); j++ {
			field := v.Type().Field(j)
			dbTag := field.Tag.Get("db")
			if dbTag == col {
				fieldValue := v.Field(j)
				if fieldValue.CanSet() {
					if values[i] == nil {
						continue
					}
					switch fieldValue.Kind() {
					case reflect.String:
						fieldValue.SetString(values[i].(string))
					case reflect.Int, reflect.Int64:
						fieldValue.SetInt(values[i].(int64))
					case reflect.Bool:
						fieldValue.SetBool(values[i].(bool))
					case reflect.Float64:
						fieldValue.SetFloat(values[i].(float64))
					case reflect.Array:
						if fieldValue.Type() == reflect.TypeOf(uuid.UUID{}) {
							if val, ok := values[i].(string); ok {
								u, err := uuid.Parse(val)
								if err == nil {
									for k := 0; k < 16; k++ {
										fieldValue.Index(k).SetUint(uint64(u[k]))
									}
								}
							}
						}
					case reflect.Struct:
						if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
							if val, ok := values[i].(time.Time); ok {
								fieldValue.Set(reflect.ValueOf(val))
							}
						}
					}
				}
				break
			}
		}
	}
	return nil
}

func join(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += sep + arr[i]
	}
	return result
}

func AutoMigrate(models ...interface{}) {
	for _, m := range models {
		v := reflect.ValueOf(m)
		if v.Kind() != reflect.Ptr {
			log.Printf("Skipping non-pointer model: %v", v)
			continue
		}

		v = v.Elem()
		if v.Kind() != reflect.Struct {
			log.Printf("Skipping non-struct model: %v", v)
			continue
		}

		log.Printf("Auto-migrating model: %v", v.Type().Name())
	}
}
