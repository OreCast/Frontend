function HideTag(tag) {
    var id=document.getElementById(tag);
    if (id) {
        id.className="hide";
    }
}
function ShowTag(tag) {
    var id=document.getElementById(tag);
    if (id) {
        id.className="show";
    }
}
function FlipTag(tag) {
    var id=document.getElementById(tag);
    if (id) {
        if  (id.className == "show") {
            id.className="hide";
        } else {
            id.className="show";
        }
    }
}
function load(url) {
    window.location.href=url;
}
function reload() {
    load(window.location.href);
}
function ShowMap() {
    // clear up orecast logo
    var id=document.getElementById('orecast-content');
    if (id) {
        id.innerHTML="";
    }
    HideTag('orecast-content');
    ShowTag('orecast-map');
}
