1) hiện tại đang tag cả tag qc-.... và stg-... sau đó mới push 1 lần

-> đổi thành push độc lập (tag qc xong push, tag stg xong push), nhưng trước khi action push -> kiểm tra thử cả 2 tags đã tồn tại chưa nha

2) mỗi project có 1 rule riêng -> suy nghĩ brain storm cách tối ưu

Nhu cầu của tôi là gõ minhthetus-cli git tag-dev-stg, nó sẽ check project hiện tại và dựa vào rule hiện tại mà cook khác nhau

có một số kiểu tag để deploy, mà tôi chiêm nghiệm đc:
- tag prefix qc để deploy sandbox: qc-v1.1.0
- tag prefix stg để deploy sandbox: stg-v1.1.0
- tag thẳng số để deploy prod: v1.0.0
- tag prefix release để deploy prod: release-v1.10.0
- tag prefix qc-cicd-v1.0.0 để deploy sandbox
