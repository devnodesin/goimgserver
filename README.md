# goimgserver

goimgserver - These services allow you to store images, then serve optimized, resized, and converted images on the fly based on URL parameters.

## Usage

1. **Setup**:

```bash
cd src
go run main.go
```

All parameters are optional: if not specified, they will start with default values:

- `--port XXXX` defaults to `9000`
- `--imagesdir /path/to/images` defaults to `{pwd}/images`
- `--cachedir /path/to/cache` defaults to `{pwd}/cache`

1. **Access the endpoints**:

```bash
curl -X GET "http://localhost:9000/img/sample.jpg/600x400"
```

- Use tools like `curl` or a browser to test the endpoints.
