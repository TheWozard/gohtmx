document.getElementById('body').addEventListener('htmx:responseError', function (evt) {
    document.getElementById('error').innerHTML = evt.detail.error
});
document.getElementById('body').addEventListener('htmx:sendError', function (evt) {
    document.getElementById('error').innerHTML = evt.detail.error
});