console.log("xxxx");
console.log("yyyyy");
postJSON("http://117.81.212.230:8025/API/record",dat);
document.getElementById('remoteSessionDescription').value= "xxxx";
console.log("yyyyy");

function postJSON(url, data, callback) {
    var request = new XMLHttpRequest();
    request.open('post', url);
    request.onreadystatechange = () => {
        if (request.readystate === 4 && callback) {
            callback(request);
        };
    }
    request.setRequestHeader('Content-Type', 'application/json');
    request.send(JSON.stringify(data))
}
