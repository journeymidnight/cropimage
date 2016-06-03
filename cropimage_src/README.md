# nginx: LPUSH taskQueue KEY

KEY format:{"uuid":"uuid-value","url":urlvalue}

# crop-image

1. download url
2. parse url's convert parameter
3. convert image based on the parameter
4. send back to nginx

# crop-image: LPUSH uuid KEY1

KEY1 format:(200|403|400)

# crop-image: SET url image-value

