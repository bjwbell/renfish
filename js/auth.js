/*jslint browser: true*/
/*globals $, FB, gapi, StackTrace, FacebookAppId, fbLogin, setSettings*/
"use strict";
function get(name) {
    var name2 = (new RegExp('[?&]' + encodeURIComponent(name) + '=([^&]*)')).exec(location.search);
    if (name2) {
        return decodeURIComponent(name2[1]);
    }
    return "";
}

function sendLogError(msg) {
    $.ajax({url: '/logerror',
            data: { 'location': window.location.href, 'error': msg },
           });
}

function logError(error) {
    var message =
        "Error: " + error + "\r\n\r\n" +
        "Stack: ",
        callback,
        errback;

    callback = function (stackframes) {
        var stringifiedStack = stackframes.map(function (sf) {
            return sf.toString();
        }).join('\n');
        message = message + stringifiedStack;
        sendLogError(message);
    };
    errback = function (err) {
        message = message + err.message;
        sendLogError(message);
    };
    StackTrace.get().then(callback, errback);
}

function gGetEmail(resp) {
    if (resp.result.emails.length === 0) {
        console.log("gRentRollCallback - error no email!");
        return '';
    }
    return resp.result.emails[0].value;
}

function gGetName(resp) {
    if (resp.result.emails.length === 0) {
        console.log("gRentRollCallback - error no email!");
        logError("gRentRollCallback - error no email!");
        return '';
    }
    return resp.result.displayName;
}

function executeCallback(name, response, logerror) {
    var callback = document.getElementById(name),
        callbackAttr = null,
        callbackName = null,
        errorMsg = '';
    if (callback === null) {
        console.log('executeCallback - ERROR NO CALLBACK: ' + name);
        if (logerror === true) {
            logError('executeCallback - ERROR NO CALLBACK: ' + name);
        }
        return;
    }
    callbackAttr = callback.attributes.callback;
    if (callbackAttr === undefined) {
        console.log('executeCallback - ERROR NO CALLBACK: ' + name);
        if (logerror === true) {
            logError('executeCallback - ERROR NO CALLBACK: ' + name);
        }
        return;
    }
    callbackName = callbackAttr.value;
    if (callbackName === undefined || callbackName === null || callbackName === '') {
        return;
    }
    // call the callback.
    if (window[callbackName] === undefined || window[callbackName] === null) {
        errorMsg =
            'executeCallback - ERROR CALLBACK DOESNT EXIST,' +
            'id name: ' + name + ', ' +
            'callback name: ' + callbackName;
        console.log(errorMsg);
        logError(errorMsg);
        return;
    }
    return window[callbackName](response);
}

function startFBLogin() {
    window.fbAsyncInit = function () {
        FB.init({
            appId      : FacebookAppId,
            xfbml      : true,
            version    : 'v2.5'
        });
        FB.getLoginStatus(function (response) {
            fbLogin(response);
        });
    };
    (function (d, s, id) {
        var js, fjs = d.getElementsByTagName(s)[0];
        if (d.getElementById(id)) {
            return;
        }
        js = d.createElement(s);
        js.id = id;
        js.src = "//connect.facebook.net/en_US/sdk.js";
        fjs.parentNode.insertBefore(js, fjs);
    }(document, 'script', 'facebook-jssdk'));
}

var gSigninFailed = 0;

function gSignin(authResult) {
    if (authResult.status.signed_in) {
        var gSigninButton = document.getElementById('gSigninButton'),
            fbSigninButton = document.getElementById('fbSigninButton'),
            gsettings = document.getElementById('gsettings');
        /*client_id = authResult.client_id, code = authResult.code,access_token = authResult.access_token,*/
        if (gSigninButton !== null) {
            gSigninButton.setAttribute('style', 'display: none');
        }
        if (fbSigninButton !== null) {
            fbSigninButton.setAttribute('style', 'display: none');
        }
        if (gsettings !== null) {
            gsettings.setAttribute('style', 'display: inline');
        }
        gapi.client.load('plus', 'v1').then(function () {
            var request = gapi.client.plus.people.get({
                'userId': 'me'
            });
            request.then(function (resp) {
                var email = "",
                    gSignoutButton = document.getElementById('gSignoutButton');
                if (resp.result.emails.length === 0) {
                    console.log("Error no email!");
                    email = "dummy@dummy.com";
                } else {
                    email = resp.result.emails[0].value;
                }
                if (gSignoutButton !== null) {
                    gSignoutButton.innerHTML = email;
                }
                executeCallback('g-callback', resp, false);
            }, function (reason) {
                console.log('gSignin - Error: ' + reason.result.error.message);
                logError('Error: ' + reason.result.error.message);
            });
        });
    } else {
        // Update the app to reflect a signed out user
        // Possible error values:
        //   "user_signed_out" - User is signed-out
        //   "access_denied" - User denied access to your app
        //   "immediate_failed" - Could not automatically log in the user
        if (gSigninFailed === 0) {
            console.log('Sign-in state: ' + authResult.error);
            startFBLogin();
            gSigninFailed += 1;
        }
    }
}

function fbLogin(response) {
    // The response object is returned with a status field that lets the
    // app know the current login status of the person.
    // Full docs on the response object can be found in the documentation
    // for FB.getLoginStatus().
    if (response.status === 'connected') {
        // Logged into your app and Facebook.
        document.getElementById('gSigninButton').setAttribute('style', 'display: none');
        var fbsettings = document.getElementById('fbsettings');
        if (fbsettings !== null) {
            fbsettings.setAttribute('style', 'display: inline');
        }
        FB.api('/me', function (response) {
            document.getElementById('fbEmail').innerHTML = response.email;
        });
        executeCallback('fb-callback', response, false);
    } else if (response.status === 'not_authorized') {
        // The person is logged into Facebook, but not your app.
        document.getElementById('fbstatus').innerHTML = 'Please log into this app.';
        executeCallback('notloggedin-callback', '', false);
    } else {
        // The person is not logged into Facebook
        executeCallback('notloggedin-callback', '', false);
    }
}

function createAccount(email, loginMethod) {
    if (email === '') {
        console.log("createAccount - empty email!");
        return;
    }
    $.ajax({url: '/createaccount',
            data: { 'email': email, 'loginmethod': loginMethod },
            success: function () {
            window.location.href = '/setup';
        }
           });
}

function gSignup(resp) {
    createAccount(gGetEmail(resp), 'Google');
}

function fbSignup() {
    FB.api('/me', function (response) {
        createAccount(response.email, 'Facebook');
    });
}

function gSettings(resp) {
    setSettings(gGetName(resp), gGetEmail(resp));
}

function fbSettings() {
    FB.api('/me', function (response) {
        setSettings(response.name, response.email);
    });

}

function setSettings(name, email) {
    var nameEle = document.getElementById("name"),
        emailEle = document.getElementById("email");
    nameEle.innerHTML = name;
    emailEle.innerHTML = email;
}

function gDummy(resp) {
}

function fbDummy() {
}
