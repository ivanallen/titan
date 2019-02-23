function Multipart() {}

Multipart.prototype._buffer = '';
Multipart.prototype._callback = {};

Multipart.prototype.setBoundary = function (boundary) {
    this._boundary = boundary;
};

Multipart.prototype.on = function (name, cb) {
    this._callback[name] = cb;
};

Multipart.prototype.write = function (chunk) {
    this._buffer += chunk;
    var index = this._buffer.search('--' + this._boundary + '--');
    if (index !== -1) {
        this._buffer = this._buffer.substring(0, index);
    }

    var parts = this._buffer.split('--' + this._boundary);
    console.log('parts:', parts)
    console.log('index:', index)

    if (index === -1) {
        // 最后一个 part 可能不完整
        this._buffer = parts.pop();
    } else {
        this._buffer = '';
    }

    for (var part of parts) {
        part = part.trim('\r\n');
        if (!part) continue;
        var res = parsePart(part);
        if (this._callback.data) {
            this._callback.data(res.headers, res.body);
        }
    }

    if (index !== -1) {
        this._callback.end(JSON.stringify({
            url: '/tmp.zip'
        }));
    }
};

function parsePart(part) {
    var t = part.split('\r\n\r\n');
    var h = t[0].split('\r\n');
    var body = t[1];
    var headers = {};
    for (var line of h) {
        var p = line.split(':');
        headers[p[0].trim()] = p[1].trim();
    }
    return {headers:headers, body:body};
}
