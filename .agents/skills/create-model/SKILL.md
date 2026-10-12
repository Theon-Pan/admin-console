---
name: create-model
description: Creates a new model based on the provided specifications.
---

# Create Model Skill
Create a new model represented by Golang struct type, the model_name `$ARGUMENTS` will be used as the name of the struct.

## Workflow

1. Connect to the posgreSQL server running on local, the connection url is defined in the `DATABASE_URL_AC` environment variable, also can be found in the `conf/admin-console.*.conf` file. Retrieve the schema information by using the model_name as the table name.
2. Generate a Golang struct type based on the retrieved schema information, using the model_name as the struct name, in Upper Camel Case, if the model_name is plural, convert it to singular form.
3. The member fields of the struct should correspond to the columns of the table, convert the column names which are named in snake case style to Upper Camel Case. And the data types should be mapped appropriately from SQL types to Go types, using the below mapping, do not use pointer except in the update struct:

| SQL Type | Go Type |
|----------|---------|
| INTEGER  | int     |
| BIGINT   | int64   |
| SERIAL   | int     |
| BIGSERIAL| int64   |
| VARCHAR  | string  |
| TEXT     | string  |
| BOOLEAN  | bool    |
| TIMESTAMP| *time.Time |
| DATE     | *time.Time |
| JSON     | []byte  |
| JSONB    | []byte  |
4. Add json tags to the struct fields, using the original column names in snake case style as the tag values.
5. Also you can ask user if they want to generate two structs for creating and updating records. The create struct should include all fields except the primary key, and the update struct should include all fields as pointers to allow partial updates.The struct names should be derived from the model_name, with `CreationRequest` and `ModificationRequest` suffixes respectively.
6. Ensure that the generated code is properly formatted according to Go conventions, including proper indentation and spacing.
7. Include necessary import statements, such as `import "time"` if any of the struct fields use the `time.Time` type.
8. Save the generated code to a file named after the singular form of model_name, with a `.go` extension, in the `internal/model` package directory.