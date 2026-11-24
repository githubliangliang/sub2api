首次真实 HTTP 夹具的 403 对照误用了 context_length_exceeded code；基线按 context 分类保留消息，所以该子例不证明普通 403 缺陷。已改用与 intake 对照相同的 invalid_encrypted_content code。Bad Request 子例真实复现主机和凭据泄漏；首次日志原样保留。
