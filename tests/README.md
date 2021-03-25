# Testing

How to test gorush with http request?

## download bat tool

Download [cURL-like tool for humans](https://github.com/astaxie/bat).

## testing

see the JSON format:

```json
{
  "notifications": [
    {
      "tokens": [
        "token_a",
        "token_b"
      ],
      "platform": 1,
      "message": "Hello World iOS!"
    },
    {
      "tokens": [
        "token_a",
        "token_b"
      ],
      "platform": 2,
      "message": "Hello World Android!"
    }
  ]
}
```

run the following command.

```sh
bat POST localhost:8088/api/push < tests/test.json
```
