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
	"newpass2":		   "z22",
	"adduser":         "zau",
	"deleteuser":      "zdu",
	"rowcount":        "zrc",
	"all":			   "xal",
	"selected":        "xse",
	"range":           "xrg",
	"savesettings":    "sss",
	"newloc":          "nlc",
	"delloc":          "ndc",
	"newcontent":      "nct",
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
	pane.className = ""
	pane.innerHTML = "";

}
function showerrormsg(msg) {

	console.log('showerrormsg '+msg);
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

	if (this.oldpass.value == '' || this.mynewpass.value == '') { 
		showerrormsg("Password must not be left blank");
		return false; 
	}
	if (this.mynewpass.value != this.mynewpass2.value) {
		showerrormsg("New passwords don't match");
		return false;
	}
	if (this.mynewpass.value.length < parseInt(this.minpwlen.value)) {
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
	let save = tr.children[4].firstElementChild;

	save.disabled = false;
	let url = "/userx?"+Param_Labels["passchg"]+"="+Param_Labels["single"];
	url += "&"+Param_Labels["userid"]+"="+uid+"&"+Param_Labels["accesslevel"]+"="+al
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			hideerrorpane();
			save.disabled = true;
		} else {
			showerrormsg(res.res);
		}
	});

}


function pwd_enableSave(inp) {

	let tr = inp.parentElement.parentElement;
	let save = tr.children[4].firstElementChild;

	console.log("Save is "+save);
	save.disabled = false;

}


function pwd_savePasswordChanges(btn) {

	let tr = btn.parentElement.parentElement;
	let uid = tr.firstElementChild.firstElementChild.value;
	let save = tr.children[4].firstElementChild;
	let np1 = tr.children[2].firstElementChild.value;
	let np2 = tr.children[3].firstElementChild.value;

	save.disabled = true;
	let url = "/userx?"+Param_Labels["passchg"]+"="+Param_Labels["single"];
	url += "&"+Param_Labels["userid"]+"="+uid
	url += "&"+Param_Labels["newpass"]+"="+encodeURIComponent(np1);
	url += "&"+Param_Labels["newpass2"]+"="+encodeURIComponent(np2);
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

function pwd_insertNewRow() {

	console.log("Inserting new user");
	let rc_obj = document.getElementById('rowcount');
	let rc = parseInt(rc_obj.value) + 1;
	let tab = document.getElementById('tabusers');
	tab.insertRow(-1);
	console.log("Row inserted, rc="+rc);
	let nr = document.getElementById('newrow');
	let nrow = tab.rows[rc];
	nrow.innerHTML = nr.innerHTML;
	
	for (let i=0; i < nrow.children.length; i++) {
		if (nrow.children[i].firstElementChild.getAttribute('name')) {
			let nn = nrow.children[i].firstElementChild.getAttribute('name')+'_'+rc;
			nrow.children[i].firstElementChild.setAttribute('name',nn);
		}
	}
	tab.rows[rc].className = "";
	rc_obj.value = rc;
	console.log('New row inserted');
	nrow.children[0].firstElementChild.focus();
}

function pwd_checkSaveNewUser() {

	let uid = document.getElementById('newuserid');
	let np1 = document.getElementById('newpass1');
	let np2 = document.getElementById('newpass2');
	let btn = document.getElementById('savenewuser');
	let ok = uid.getAttribute('data-ok') == '1'
	ok = ok && np1.getAttribute('data-ok') == '1';
	ok = ok && np1.value == np2.value;
	btn.disabled = !ok;
	
}

function pwd_useridChanged(obj) {

	let uid = obj.value;
	let row = obj.getAttribute('name').match(/_(\d+)/);
	let tab = obj.parentElement.parentElement.parentElement;
	if (obj.value.length > 0) {
		obj.setAttribute('data-ok','1');
	}
	for (let r = 0; r < tab.rows.length; r++) {
		let ruid = tab.rows[r].firstElementChild;
		//console.log(ruid.innerHTML+" == "+uid);
		ruid = ruid.firstElementChild;
		obj.classList.remove('warning');
		if (r + 1 == row[1]) continue;
		if (ruid.value == uid) {
			console.log("Error - duplicate! "+row+"/"+uid+" == "+r+"/"+ruid.value);
			obj.classList.add('warning');
			obj.setAttribute('data-ok','0');
			break;
		}
	}
	pwd_checkSaveNewUser();

}

function pwd_checkpass(inp) {

	let pwl = document.getElementById('minpwlen');
	if (inp.value.length >= pwl.value) {
		inp.setAttribute('data-ok','1');
	} else {
		inp.setAttribute('data-ok','0');
	}
	pwd_checkSaveNewUser();

}

function pwd_insertNewUser(btn) {

	btn.disabled = true;

	let uid = document.getElementById('newuserid').value;
	let al = document.getElementById('newal').value;
	let np1 = document.getElementById('newpass1').value;
	let np2 = document.getElementById('newpass2').value;

	let url = "/userx?"+Param_Labels["adduser"]+"="+uid;
	url += "&"+Param_Labels["accesslevel"]+"="+al
	url += "&"+Param_Labels["newpass"]+"="+encodeURIComponent(np1);
	url += "&"+Param_Labels["newpass2"]+"="+encodeURIComponent(np2);
	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			console.log("Fetching users");
			window.location.replace("/users");
		} else {
			showerrormsg(res.res);
		}
	});

}

function param_select_keys(selectall,key) {

	let divs = document.querySelectorAll("#"+key+"filter>.filteritems");
	for (let i = 0; i < divs.length; i++) {
		if (selectall) {
			divs[i].classList.add('hide');
		} else {
			divs[i].classList.remove('hide');
		}
	}
	let cbs = document.getElementsByName(Param_Labels[key]);
	for (let i = 0; i < cbs.length; i++) {
		cbs[i].checked = false;
	}
	enableSaveSettings();
}

function param_select_locations(selectall) {

	param_select_keys(selectall,'location');

}

function param_select_owners(selectall) {

	param_select_keys(selectall,'owner');

}

function enableSaveSettings() {

	let cmd = document.getElementById('savesettings');
	if (!cmd) return;
	cmd.disabled = false;

}
function trapDirtyPage() {

	// This method does not allow for clearing lock flags when definitely leaving a dirty page
	
	
	window.addEventListener('beforeunload', function(e) {
	
	var cmd = document.getElementById('savesettings');
	if (cmd == null)
		cmd = document.getElementById('savedata'); 
	if (cmd == null)
		return;
	var myPageIsDirty = !cmd.disabled && cmd.getAttribute('data-triggered')=='0';  //you implement this logic...
	if (myPageIsDirty) {
		//following two lines will cause the browser to ask the user if they
		//want to leave. The text of this dialog is controlled by the browser.
		e.preventDefault(); //per the standard
		e.returnValue = ''; //required for Chrome
	}
		//else: user is allowed to leave without a warning dialog
	});
}

function bodyLoaded() {

	trapkeys();
	trapDirtyPage();

}

function add_new_location(obj) {

	obj.disabled = true;

	let tr = obj.parentElement.parentElement;
	let newloc = tr.firstElementChild.firstElementChild.value;
	let url = "/locations?"+Param_Labels["newloc"]+"="+encodeURIComponent(newloc)
	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			console.log("Fetching locations");
			window.location.replace("/locations");
		} else {
			obj.disabled = false;
			showerrormsg(res.res);
		}
	});

}


function delete_location(obj) {

	obj.disabled = true;

	let tr = obj.parentElement.parentElement;
	let loc = tr.firstElementChild.firstElementChild.innerText
	let url = "/locations?"+Param_Labels["delloc"]+"="+encodeURIComponent(loc)
	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			console.log("Fetching locations");
			window.location.replace("/locations");
		} else {
			obj.disabled = false;
			showerrormsg(res.res);
		}
	});

}

function add_new_box_content(obj) {

	obj.disabled = true;

	let tr = obj.parentElement.parentElement;
	let box = obj.getAttribute('data-boxid');
	let owner = tr.children[0].firstElementChild.value.toLocaleUpperCase();
	let client = tr.children[1].firstElementChild.value.toLocaleUpperCase();
	let name = tr.children[2].firstElementChild.value;
	let contents = tr.children[3].firstElementChild.value;
	let review = tr.children[4].firstElementChild.value;


	let url = "/boxes?"+Param_Labels["newcontent"]+"="+encodeURIComponent(box);
	url += "&"+Param_Labels["owner"]+"="+encodeURIComponent(owner);
	url += "&"+Param_Labels["client"]+"="+encodeURIComponent(client);
	url += "&"+Param_Labels["name"]+"="+encodeURIComponent(name);
	url += "&"+Param_Labels["contents"]+"="+encodeURIComponent(contents);
	url += "&"+Param_Labels["review_date"]+"="+encodeURIComponent(review);
	
	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			window.location.replace("/boxes?"+Param_Labels["boxid"]+"="+encodeURIComponent(box));
		} else {
			obj.disabled = false;
			showerrormsg(res.res);
		}
	});

}

function fetch_client_name_list(obj) {

	let client = obj.value;
	let url = "/boxes?"+Param_Labels["client"]+"="+encodeURIComponent(client);

	console.log(url);
	fetch(url)
	.then(res => res.json())
	.then(res => {
		if (res.res != "ok") return;
		console.log(res.names);
		let dl = document.getElementById("namelist");
		dl.textContent = "";
		res.names.forEach(element => {
			let opt = document.createElement("option");
			opt.value = element;
			dl.appendChild(opt);
		});
	})
}

