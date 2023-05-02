const Param_Labels = {
	"boxid":           "qbx",
	"owner":           "qow",
	"contents":        "qcn",
	"review_date":     "qdt",
	"name":            "qnm",
	"client":          "qcl",
	"location":        "qlo",
	"numdocs":         "qnd",
	"min_review_date": "qd1",
	"max_review_date": "qd2",
	"userid":          "quu", // Hardcoded in boxes.js!
	"userpass":        "qup",
	"accesslevel":     "qal", // Hardcoded in boxes.js!
	"pagesize":        "qps", // Hardcoded in boxes.js!
	"offset":          "qof",
	"order":           "qor",
	"find":            "qqq",
	"desc":            "qds",
	"field":           "qfd",
	"overview":        "qov",
	"table":           "qtb",
	"textfile":        "qtx",
	"passchg":         "zpc", // Hardcoded in boxes.js!
	"single":          "z11", // Hardcoded in boxes.js!
	"multiple":        "z99",
	"oldpass":         "zop",
	"newpass":         "znp",
	"deleteuser":      "zdu",
	"rowcount":        "zrc",
}


function isBadLength(sObj,iLen,sMsg) {

    if (sObj.value.length < iLen) {
      alert(sMsg)
      sObj.focus()
      return true
    }
  }

function changepagesize(sel) {
	let newpagesize = sel.value;
	let url = window.location.href;
	// Need to strip out any existing PAGESIZE
	let ps = url.match(/(&|\?)qps\=\d+/);   // qps must match Param_Labels["pagesize"]
	console.log('url="'+url+'"; ps="'+ps+'"; NP='+newpagesize);
	let cleanurl = url;
	if (ps) {
		cleanurl = cleanurl.replace(ps[0],'') + ps[1];
	} else {
		if (cleanurl.indexOf('?') < 0) {
			cleanurl += '?';
		} else {
			cleanurl += '&';
		}
	}
	console.log("cleanurl='"+cleanurl+"'");
	window.location.href = cleanurl + "qps=" + newpagesize;
}
function trapkeys() {
	document.getElementsByTagName('body')[0].onkeyup = function(e) { 
		var ev = e || window.event;
	 	if (ev.keyCode == 37 || ev.keyCode == 33) { // Left arrow or PageUp
	   		let pp = document.getElementById('prevpage');
			if (pp) {
				window.location.href = pp.getAttribute('href');
			}
	   		return false;
		} else if (ev.keyCode == 39 || ev.keyCode == 34) { // Right arrow or PageDn
			let np = document.getElementById('nextpage');
		 	if (np) {
				window.location.href = np.getAttribute('href');
		 	}
			return false;
	    } 
	}

	let el = document.querySelector('body');
	swipedetect(el, function(swipedir){
		/* swipedir contains either "none", "left", "right", "top", or "down" */
		if (swipedir =='left') {
			console.log("swiped left");
			let pp = document.getElementById('prevpage');
			if (pp) {
				window.location.href = pp.getAttribute('href');
			}
		}
		else if (swipedir =='right') {
			alert("swiped right");
			let pp = document.getElementById('nextpage');
			if (pp) {
				window.location.href = pp.getAttribute('href');
			}
		}

	})
	



}

function activatemsgpane(msg,cssclass) {

	let pane = document.getElementById('errormsgdiv');
	if (!pane) { return; }
	pane.classList.add(cssclass);
	pane.innerHTML = msg;
}

function hideerrorpane() {

	let pane = document.getElementById('errormsgdiv');
	if (!pane) { return; }
	pane.class = ""
	pane.innerHTML = "msg";

}
function showerrormsg(msg) {

	activatemsgpane(msg,"errormsg");
	
}

function showwarning(msg) {

	activatemsgpane(msg,"warning");
	
}

function swipedetect(el, callback){
  
    var touchsurface = el,
    swipedir,
    startX,
    startY,
    distX,
    distY,
    threshold = 150, //required min distance traveled to be considered swipe
    restraint = 100, // maximum distance allowed at the same time in perpendicular direction
    allowedTime = 300, // maximum time allowed to travel that distance
    elapsedTime,
    startTime,
    handleswipe = callback || function(swipedir){}
  
    touchsurface.addEventListener('touchstart', function(e){
        var touchobj = e.changedTouches[0]
        swipedir = 'none'
        dist = 0
        startX = touchobj.pageX
        startY = touchobj.pageY
        startTime = new Date().getTime() // record time when finger first makes contact with surface
        e.preventDefault()
    }, false)
  
    touchsurface.addEventListener('touchmove', function(e){
        e.preventDefault() // prevent scrolling when inside DIV
    }, false)
  
    touchsurface.addEventListener('touchend', function(e){
        var touchobj = e.changedTouches[0]
        distX = touchobj.pageX - startX // get horizontal dist traveled by finger while in contact with surface
        distY = touchobj.pageY - startY // get vertical dist traveled by finger while in contact with surface
        elapsedTime = new Date().getTime() - startTime // get time elapsed
        if (elapsedTime <= allowedTime){ // first condition for awipe met
            if (Math.abs(distX) >= threshold && Math.abs(distY) <= restraint){ // 2nd condition for horizontal swipe met
                swipedir = (distX < 0)? 'left' : 'right' // if dist traveled is negative, it indicates left swipe
            }
            else if (Math.abs(distY) >= threshold && Math.abs(distX) <= restraint){ // 2nd condition for vertical swipe met
                swipedir = (distY < 0)? 'up' : 'down' // if dist traveled is negative, it indicates up swipe
            }
        }
        handleswipe(swipedir)
        e.preventDefault()
    }, false)
}
  
//USAGE:
/*
var el = document.getElementById('someel')
swipedetect(el, function(swipedir){
    swipedir contains either "none", "left", "right", "top", or "down"
    if (swipedir =='left')
        alert('You just swiped left!')
})
*/



// Password maintenance stuff

function pwd_validateSingleChange(frm) {

	if (this.oldpass.value == '' || this.newpass.value == '') { 
		showerrormsg("Password must not be left blank");
		return false; 
	}
	if (this.newpass.value != this.newpass2.value) {
		showerrormsg("New passwords don't match");
		return false;
	}
	if (this.newpass.value.length < parseInt(this.minpwlen.value)) {
		showerrormsg("Password not long enough");
		return false;
	}
	return true;
}

function pwd_deleteUser(btn) {

	let tr = btn.parentElement.parentElement;
	let tab = tr.parentElement;
	let uid = tr.firstElementChild.firstElementChild.value;

	let url = "/userx?"+Param_Labels["passchg"]+"="+Param_Labels["single"];
	url += "&"+Param_Labels["userid"]+"="+uid
	url += "&"+Param_Labels["deleteuser"]+"="+Param_Labels["deleteuser"]
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(function (res) {
		console.log(res.res);
		if (res.res=="ok") {
			hideerrorpane();
			console.log("row is "+tr.rowIndex);
			tab.removeChild(tr);
		} else {
			showerrormsg(res.res);
		}
	});

}

function pwd_updateAccesslevel(sel) {

	let al = sel.value;
	let tr = sel.parentElement.parentElement;
	let uid = tr.firstElementChild.firstElementChild.value;

	let url = "/userx?"+Param_Labels["passchg"]+"="+Param_Labels["single"];
	url += "&"+Param_Labels["userid"]+"="+uid+"&"+Param_Labels["accesslevel"]+"="+al
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			hideerrorpane();
		} else {
			showerrormsg(res.res);
		}
	});

}