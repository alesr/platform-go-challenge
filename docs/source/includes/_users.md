# Users

## List Users

```shell
curl "http://localhost:8090/users"
```

> The above command returns JSON structured like this:

```json
{
  "status": "success",
  "data": {
    "data": [
      {
        "id": "01JM9R7XTFBS7QBDZCV9V6HC4Q",
        "name": "Mr. Kim Franecki",
        "created_at": "2267-06-05T19:45:59.227296952Z",
        "updated_at": "2256-04-29T03:58:36.788817227Z"
      },
      {
        "id": "01JM9R7XTHP89ZW3GF1E790TJA",
        "name": "King Austin Lehner",
        "created_at": "2033-10-11T15:02:46.319856777Z",
        "updated_at": "2213-12-31T21:00:43.970100209Z"
      },
      {
        "id": "01JM9R7XTEB1FFKW5B22YKTYYS",
        "name": "Lord Bradley Wisoky",
        "created_at": "2220-08-01T09:58:16.329011286Z",
        "updated_at": "2052-05-13T11:56:39.433896664Z"
      }
    ]
  }
}
```

This endpoint retrieves a list of test users that can be used for marking assets as favorites and updating them.

### HTTP Request

`GET http://localhost:8090/users`

### Response Fields

Field | Type | Description
--------- | ---- | -----------
id | string | Unique identifier for the user
name | string | User's name
created_at | string | Timestamp of when the user was created
updated_at | string | Timestamp of when the user was last updated
