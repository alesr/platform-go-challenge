# Errors

The Platform Go Challenge API uses the following error codes:

Error Code | Meaning
---------- | -------
400 | Bad Request -- Invalid request parameters or payload:<br>• Invalid page size<br>• Invalid maximum results value<br>• Invalid page token<br>• Invalid favorite asset payload<br>• Invalid user ID<br>• Invalid favorite ID<br>• Invalid asset ID<br>• Description too long<br>• Missing required user ID<br>• Missing required favorite ID<br>• Unsupported asset type
404 | Not Found -- The specified resource could not be found:<br>• User not found<br>• Asset not found<br>• Favorite asset not found
500 | Internal Server Error:<br>• We had a problem with our server<br>• Invalid data in storage


All errors are returned in the following format:

```json
{
  "status-code": 404,
  "message": "Favorite asset not found"
}
```
