document.getElementById('body').addEventListener('htmx:responseError', function (evt) {
    console.log(evt)
    document.getElementById('error').innerHTML = evt.detail.error
});
document.getElementById('body').addEventListener('htmx:sendError', function (evt) {
    console.log(evt)
    document.getElementById('error').innerHTML = evt.detail.error
});