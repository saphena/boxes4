const Param_Labels = {
	"boxid":           "qbx",
	"owner":           "qow",
	"contents":        "qcn",
	"review_date":     "qdt",
	"name":            "qnm",
	"client":          "qcl",
	"location":        "qlo",
	"storeref":        "qlr",
	"numdocs":         "qnd",
	"min_review_date": "qd1",
	"max_review_date": "qd2",
	"userid":          "quu", 
	"userpass":        "qup",
	"accesslevel":     "qal", 
	"pagesize":        "qps", 
	"offset":          "qof",
	"order":           "qor",
	"find":            "qqq",
	"desc":            "qds",
	"field":           "qfd",
	"overview":        "qov",
	"table":           "qtb",
	"textfile":        "qtx",
	"passchg":         "zpc", 
	"single":          "z11", 
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
	"delcontent":      "dct",
	"savecontent":     "dsc",
	"chgboxlocn":      "dxl",
	"savebox":         "dbx",
	"newok":           "xid",
	"newbox":          "xnb",
	"delbox":          "kbx",
	"ExcludeBeforeYear": "xby",
	"theme":			"ttt",
	"delowner":			"kwn",
}



function changePagesize(sel) {
	let newpagesize = sel.value;

	setPagesize(newpagesize);

}

function loadPrevPage() {
	let pp = document.getElementById('prevpage');
	if (pp) {
		window.location.href = pp.getAttribute('href');
	}
	   return false;
}

function loadNextPage() {
	let np = document.getElementById('nextpage');
	if (np) {
	   window.location.href = np.getAttribute('href');
	}
   return false;
}
function trapKeys() {
	document.getElementsByTagName('body')[0].onkeyup = function(e) { 
		var ev = e;
	 	if (ev.key == "ArrowLeft" || ev.key == "PageUp") { // Left arrow or PageUp
			return loadPrevPage();
		} else if (ev.key == "ArrowRight" || ev.key == "PageDown") { // Right arrow or PageDn
			return loadNextPage();
	    } 
	}
}

function activateMsgPane(msg,cssclass) {

	let pane = document.getElementById('errormsgdiv');
	if (!pane) { return; }
	pane.classList.add(cssclass);
	pane.innerHTML = msg;
}

function hideErrorPane() {

	let pane = document.getElementById('errormsgdiv');
	if (!pane) { return; }
	pane.className = "hide"
	pane.innerHTML = "";

}

function showErrorMsg(msg) {

	console.log('showErrorMsg '+msg);
	activateMsgPane(msg,"errormsg");
	
}

function showWarning(msg) {

	activateMsgPane(msg,"warning");
	
}

let touchstartX = 0
let touchendX = 0
let touchthresholdX = 20; // Arbitrary value chosen using a Lenovo tablet

function checkDirection() {
  let delta = touchendX - touchstartX;
  if (Math.abs(delta) < touchthresholdX) return;
  if (touchendX < touchstartX ) { // swipe left
//	alert("" + (touchendX - touchstartX));
	loadNextPage();
  }
  if (touchendX > touchstartX) {
	loadPrevPage();
  }
  
  
}

document.addEventListener('touchstart', e => {
  touchstartX = e.changedTouches[0].screenX
})

document.addEventListener('touchend', e => {
  touchendX = e.changedTouches[0].screenX
  checkDirection()
})




// Password maintenance stuff

function pwd_validateSingleChange(frm) {

	if (this.oldpass.value == '' || this.mynewpass.value == '') { 
		showErrorMsg("Password must not be left blank");
		return false; 
	}
	if (this.mynewpass.value != this.mynewpass2.value) {
		showErrorMsg("New passwords don't match");
		return false;
	}
	if (this.mynewpass.value.length < parseInt(this.minpwlen.value)) {
		showErrorMsg("Password not long enough");
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
			hideErrorPane();
			console.log("row is "+tr.rowIndex);
			tab.removeChild(tr);
		} else {
			showErrorMsg(res.res);
		}
	});

}

function pwd_updateAccesslevel(sel) {

	const savebutton = 4;

	let al = sel.value;
	let tr = sel.parentElement.parentElement;
	let uid = tr.firstElementChild.firstElementChild.value;
	let save = tr.children[savebutton].firstElementChild;

	save.disabled = false;
	let url = "/userx?"+Param_Labels["passchg"]+"="+Param_Labels["single"];
	url += "&"+Param_Labels["userid"]+"="+uid+"&"+Param_Labels["accesslevel"]+"="+al
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			hideErrorPane();
			save.disabled = true;
		} else {
			showErrorMsg(res.res);
		}
	});

}


function pwd_enableSave(inp) {

	const savebutton = 4;

	let tr = inp.parentElement.parentElement;
	let save = tr.children[savebutton].firstElementChild;

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
			hideErrorPane();
		} else {
			showErrorMsg(res.res);
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
	let ids = nr.getAttribute('data-fields').split(',');
	nrow.innerHTML = nr.innerHTML;
	nrow.id = '';
	for (let i=0; i < nrow.children.length; i++) {
		if (i < ids.length) {
			nrow.children[i].firstElementChild.setAttribute('id',ids[i]);
		}
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
	for (let r = 0; r < tab.rows.length - 1; r++) {
		let ruid = tab.rows[r].firstElementChild;
		console.log(ruid.innerHTML+" == "+uid);
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
			showErrorMsg(res.res);
		}
	});

}

function param_selectKeys(selectall,key) {

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

function param_selectDates(selectall) {

	let dtls = document.getElementById('daterangedetails');
	if (!dtls) return;
	if (selectall) {
		dtls.classList.add('hide');
	} else {
		dtls.classList.remove('hide');
	}
	enableSaveSettings();
}

function param_selectLocations(selectall) {

	param_selectKeys(selectall,'location');

}

function param_selectOwners(selectall) {

	param_selectKeys(selectall,'owner');

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

	trapKeys();
	trapDirtyPage();

}

function addNewLocation(obj) {

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
			showErrorMsg(res.res);
		}
	});

}


function deleteLocation(obj) {

	obj.disabled = true;

	let tr = obj.parentElement.parentElement;
	let loc = tr.firstElementChild.innerText
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
			showErrorMsg(res.res);
		}
	});

}

function checkNewBoxid(obj) {

	let boxid = obj.value.toLocaleUpperCase();
	
	let btn = obj.nextElementSibling;
	btn.disabled = true;
	if (boxid.length < 1) return;

	let url = "/boxes?"+Param_Labels["newok"]+"="+encodeURIComponent(boxid)

	fetch(url,{method: "GET"})
	.then(res => res.json())
	.then(res => {
		console.log(res);
		if (res.res=="ok") {
			console.log(btn.value);
			btn.disabled = false
		}
	});

}

function addNewBoxContent(obj) {

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

			let tab = tr.parentElement;
			if (tab.rows.length < 2) { // This is the first line in the box, nothing available to clone so reload
				window.location.replace("/boxes?"+Param_Labels["boxid"]+"="+encodeURIComponent(box));
			}
			// Add this line at the top of the list
			let newRow = tab.insertRow(1);
			newRow.setAttribute('data-id',res.recid)
			let newOwner = newRow.insertCell();
			let newClient = newRow.insertCell();
			let newName = newRow.insertCell();
			let newContents = newRow.insertCell();
			let newReview = newRow.insertCell();
			let newButtons = newRow.insertCell();

			newOwner.innerText = owner;
			newOwner.setAttribute('contenteditable','true');
			newOwner.setAttribute('oninput','contentSaveNeeded(this);');
			newClient.innerText = client;
			newClient.setAttribute('contenteditable','true');
			newClient.setAttribute('oninput','contentSaveNeeded(this);');
			newName.innerText = name;
			newName.setAttribute('contenteditable','true');
			newName.setAttribute('oninput','contentSaveNeeded(this);');
			newContents.innerText = contents;
			newContents.setAttribute('contenteditable','true');
			newContents.setAttribute('oninput','contentSaveNeeded(this);');
			newReview.innerHTML = tr.children[4].innerHTML;
			newReview.firstElementChild.setAttribute('onchange','contentSaveNeeded(this.parentElement);');
			newButtons.innerHTML = tab.rows[2].children[5].innerHTML;

			// Now reset the input line

			tr.children[0].firstElementChild.value = "";
			tr.children[1].firstElementChild.value = "";
			tr.children[2].firstElementChild.value = "";
			tr.children[3].firstElementChild.value = "";

			tr.children[0].firstElementChild.focus();

		} else {
			obj.disabled = false;
			console.log('ShowingErrorMsg')
			showErrorMsg(res.res);
		}
	});

}

function fetchClientNamelist(obj) {

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
		let n = 0;
		let v = "";
		res.names.forEach(element => {
			n++;
			let opt = document.createElement("option");
			v = element;
			opt.value = element;
			dl.appendChild(opt);
		});
		if (n == 1) {
			let tr = obj.parentElement.parentElement;
			let name = tr.children[2].firstElementChild;
			name.value = v; 
		}
	})
}

function deleteBoxContentLine(obj) {

	obj.disabled = true;

	let id = obj.getAttribute('data-id');
	let boxid = obj.getAttribute('data-boxid');
	let url = "/boxes?"+Param_Labels["delcontent"]+"="+id;

	// Delete is dangerous so let's do some belt and braces
	let tr = obj.parentElement.parentElement;
	let owner = tr.children[0].innerText
	let client = tr.children[1].innerText

	url += "&"+Param_Labels["owner"]+"="+encodeURIComponent(owner)
	url += "&"+Param_Labels["client"]+"="+encodeURIComponent(client)
	url += "&"+Param_Labels["boxid"]+"="+encodeURIComponent(boxid)

	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			window.location.replace("/boxes?"+Param_Labels["boxid"]+"="+encodeURIComponent(boxid));
		} else {
			obj.disabled = false;
			showErrorMsg(res.res);
		}
	});

}

// returns the save button object on the current line
// this points to a contenteditable TD
function my_UBC_button(obj) {

	console.log('obj: '+JSON.stringify(obj));
	let tr = obj.parentElement;
	let btn = tr.children[5].firstElementChild;
	return btn;

}

function autosave_OwnerName(obj) {

	obj.classList.add('warning');

	let btn = document.getElementById('saveOwnerName');
	btn.classList.remove('hide');
	btn.disabled = false;
	let ass = 0;
	let assobj = document.getElementById('AutosaveSeconds');
	if (assobj) ass = assobj.value;
	console.log('ass: '+ass);
	if (btn.timer) {
		clearTimeout(btn.timer);
	}
	btn.timer = setTimeout(updateOwnerName,ass * 1000,btn);

}

function updateOwnerName(btn) {

	let namefield = document.getElementById('ownername');
	btn.disabled = true;
	let owner = namefield.getAttribute('data-owner');
	let name = namefield.value;
	if (name == '') name = ' '; // an empty string won't be detected in Go

	let url = "/owners?"+Param_Labels["owner"]+"="+owner;
	url += "&"+Param_Labels["name"]+"="+encodeURIComponent(name);

	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			namefield.classList.remove('warning');
			btn.classList.add('hide');
		} else {
			btn.disabled = false;
			showErrorMsg(res.res);
		}
	});

}


function deleteChildlessOwner(btn) {

	btn.disabled = true;
	let owner = btn.getAttribute('data-owner');

	let url = "/owners?"+Param_Labels["owner"]+"="+owner;
	url += "&"+Param_Labels["delowner"]+"=1";

	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			btn.classList.add('hide');
			window.location.replace("/owners");
		} else {
			btn.disabled = false;
			showErrorMsg(res.res);
		}
	});

}


function autosave_UBC(obj) {

	let btn = my_UBC_button(obj);
	let ass = 0;
	let assobj = document.getElementById('AutosaveSeconds');
	if (assobj) ass = assobj.value;
	console.log('ass: '+ass);
	if (btn.timer) {
		clearTimeout(btn.timer);
	}
	btn.timer = setTimeout(updateBoxContentLine,ass * 1000,btn);

}

function autosave_Box() {

	let btn = document.getElementById('updateboxbutton');
	let ass = 0;
	let assobj = document.getElementById('AutosaveSeconds');
	if (assobj) ass = assobj.value;
	console.log('ass: '+ass);
	if (btn.timer) {
		clearTimeout(btn.timer);
	}
	btn.timer = setTimeout(updateBoxDetails,ass * 1000,btn);

}
function boxDetailsSaved() {

	alert('boxDetailsSaved');

}


function updateBoxDetails(obj) {

	obj.disabled = true;
	let boxid = obj.getAttribute('data-boxid');
	let storeref = document.getElementById('boxstoreref');
	let overview = document.getElementById('boxoverview');

	let url = "/boxes?"+Param_Labels["savebox"]+"="+boxid;
	url += "&"+Param_Labels["storeref"]+"="+encodeURIComponent(storeref.innerText);
	url += "&"+Param_Labels["overview"]+"="+encodeURIComponent(overview.innerText);

	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			storeref.classList.remove('warning');
			overview.classList.remove('warning');
			obj.classList.add('hide');
		} else {
			obj.disabled = false;
			showErrorMsg(res.res);
		}
	});


}

function updateBoxContentLine(obj) {

	obj.disabled = true;

	let id = obj.getAttribute('data-id');
	let boxid = obj.getAttribute('data-boxid');
	let url = "/boxes?"+Param_Labels["savecontent"]+"="+id;

	let tr = obj.parentElement.parentElement;
	let owner = tr.children[0].innerText
	let client = tr.children[1].innerText
	let name = tr.children[2].innerText
	let contents = tr.children[3].innerText
	let review = tr.children[4].firstElementChild.value;

	url += "&"+Param_Labels["owner"]+"="+encodeURIComponent(owner)
	url += "&"+Param_Labels["client"]+"="+encodeURIComponent(client)
	url += "&"+Param_Labels["name"]+"="+encodeURIComponent(name)
	url += "&"+Param_Labels["contents"]+"="+encodeURIComponent(contents)
	url += "&"+Param_Labels["review_date"]+"="+encodeURIComponent(review)
	url += "&"+Param_Labels["boxid"]+"="+encodeURIComponent(boxid)

	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			contentNowSaved(tr);
			let nf = document.getElementById('boxnumfiles');
			let dt = document.getElementById('boxdates');
			if (nf) {
				nf.innerText = res.nfiles;
			}
			if (dt) {
				if (res.lodate==res.hidate) {
					dt.innerText = res.lodate;
				} else {
					dt.innerText = res.lodate+' to '+res.hidate;
				}
			}
		} else {
			obj.disabled = false;
			showErrorMsg(res.res);
		}
	});

}

// called from a SELECT holding month/year assuming container
// has input, select, select holding iso8601, mm, yyyy respectively
function date_from_selects(obj) {

	let con = obj.parentElement;
	let dt = con.children[0];
	let mm = con.children[1];
	let yy = con.children[2];

	dt.value = yy.value+"-"+mm.value+"-01"
	dt.onchange();

}

function newContentSaveNeeded(obj) {

	const savebutton = 5;
	obj.classList.add('warning');
	let tr = obj.parentElement.parentElement;
	let save = tr.children[savebutton].firstElementChild; 
	let ok = true;
	for (let i = 0; i < savebutton; i++) {
		ok = ok && tr.children[i].firstElementChild.value != "";
	}
	save.disabled = !ok;

}

function boxDetailsSaveNeeded(obj) {

	obj.classList.add('warning');
	save = document.getElementById('updateboxbutton');
	save.classList.remove('hide');
	autosave_Box();

}

// Called when a box content record is edited
function contentSaveNeeded(obj) {

	autosave_UBC(obj);
	obj.classList.add('warning');
	let tr = obj.parentElement;
	console.log('Found TR');
	let save = tr.children[5].firstElementChild;

	// There might not be a second button so preserve the order below
	let del = tr.children[5].lastElementChild;

	save.classList.remove('hide');
	del.classList.add('hide');
	
}


// Save is now complete so clean the whole line
function contentNowSaved(tr) {

	const savebutton = 5;

	let save = tr.children[savebutton].firstElementChild;

	// There might not be a second button so preserve the order below
	let del = tr.children[savebutton].lastElementChild;
	for (let i = 0; i < savebutton; i++) {
		tr.children[i].classList.remove('warning');
	}
	save.classList.add('hide');
	save.disabled = false;
	del.classList.remove('hide');

}

function changeBoxLocation(sel) {

	sel.classList.add('warning');
	let loc = sel.value;
	let boxid = document.getElementById('boxboxid').innerText;
	let url = "/boxes?"+Param_Labels["chgboxlocn"]+"="+encodeURIComponent(loc)
	url += "&"+Param_Labels["boxid"]+"="+encodeURIComponent(boxid)
	console.log(url);
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			sel.classList.remove('warning');
		} else {
			sel.disabled = false;
			showErrorMsg(res.res);
		}
	});


}

function deleteEmptyBox(boxid) {

	let url = "/boxes?"+Param_Labels["delbox"]+"="+encodeURIComponent(boxid)
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			url = "/boxes"
			window.location.replace(url)
		} else {
			showErrorMsg(res.res);
		}
	});
}



function startNewBox(btn) {

	btn.disabled = true;
	let tr = btn.parentElement.parentElement;
	let boxid = tr.children[0].firstElementChild.value;
	boxid = boxid.toLocaleUpperCase();
	let url = "/boxes?"+Param_Labels["newbox"]+"="+encodeURIComponent(boxid)
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			url = "/boxes?"+Param_Labels["boxid"]+"="+encodeURIComponent(boxid)
			window.location.replace(url)
		} else {
			btn.disabled = false;
			showErrorMsg(res.res);
		}
	});

}

function setPagesize(pagesize) {

	console.log('setPagesize '+pagesize)
	let url = "/ps?"+Param_Labels["pagesize"]+"="+encodeURIComponent(pagesize)
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			console.log(window.location);
			window.location.replace(window.location)
		} else {
			showErrorMsg(res.res);
		}
	});

}

function setTheme(theme) {

	console.log('setTheme '+theme)
	let url = "/theme?"+Param_Labels["theme"]+"="+encodeURIComponent(theme)
	fetch(url,{method: "POST"})
	.then(res => res.json())
	.then(res => {
		if (res.res=="ok") {
			console.log(window.location);
			window.location.replace(window.location)
		} else {
			showErrorMsg(res.res);
		}
	});

}
