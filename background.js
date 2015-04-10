// Generated by LiveScript 1.3.1
(function(){
  var root, postJsonExt, getCookie, getRemoteCookies, addlogfb, addlog;
  root = typeof exports != 'undefined' && exports !== null ? exports : this;
  postJsonExt = function(url, jsondata, callback){
    return $.ajax({
      type: 'POST',
      url: url,
      data: JSON.stringify(jsondata),
      success: function(data){
        if (callback != null) {
          return callback(data);
        }
      },
      contentType: 'application/json'
    });
  };
  getCookie = function(callback){
    return chrome.cookies.getAll({
      url: baseurl + '/'
    }, function(cookie){
      var output, i$, len$, x, name, value;
      output = {};
      for (i$ = 0, len$ = cookie.length; i$ < len$; ++i$) {
        x = cookie[i$];
        name = decodeURIComponent(x.name);
        value = decodeURIComponent(x.value);
        output[name] = value;
      }
      return callback(output);
    });
  };
  getRemoteCookies = function(username, callback){
    if (username == null || username === 'Anonymous User' || username.length === 0) {
      return;
    }
    return $.getJSON(baseurl + '/cookiesforuser?' + $.param({
      username: username
    }), function(cookies){
      var yearlater, k, v;
      yearlater = Math.floor(Date.now() / 1000.0 + 3600 * 24 * 365);
      for (k in cookies) {
        v = cookies[k];
        chrome.cookies.set({
          url: baseurl + '/',
          name: k,
          value: encodeURIComponent(v.toString()),
          path: '/',
          expirationDate: yearlater
        });
      }
      if (root.fbname != null && root.fbname.length > 0) {
        chrome.cookies.set({
          url: baseurl + '/',
          name: 'fbname',
          value: encodeURIComponent(root.fbname.toString()),
          path: '/',
          expirationDate: yearlater
        });
      }
      if (root.fburl != null && root.fburl.length > 0) {
        chrome.cookies.set({
          url: baseurl + '/',
          name: 'fburl',
          value: encodeURIComponent(root.fburl.toString()),
          path: '/',
          expirationDate: yearlater
        });
      }
      return callback(cookies);
    });
  };
  addlogfb = function(logdata, cookie){
    var data, ref$;
    data = $.extend({}, logdata);
    data.username = (ref$ = cookie.fullname) != null
      ? ref$
      : root.fbname;
    data.lang = cookie.lang;
    data.format = cookie.format;
    data.scriptformat = cookie.scriptformat;
    data.time = Date.now();
    data.timeloc = new Date().toString();
    return postJsonExt(baseurl + '/addlogfb', data);
  };
  addlog = function(logdata, cookie){
    var data, ref$;
    data = $.extend({}, logdata);
    data.username = (ref$ = cookie.fullname) != null
      ? ref$
      : root.fbname;
    data.lang = cookie.lang;
    data.format = cookie.format;
    data.scriptformat = cookie.scriptformat;
    data.time = Date.now();
    data.timeloc = new Date().toString();
    return postJsonExt(baseurl + '/addlog', data);
  };
  root.fbname = null;
  root.fburl = null;
  root.sentmissingcookie = false;
  root.sentmissingformat = false;
  chrome.runtime.onMessage.addListener(function(request, sender, sendResponse){
    var fbname, fburl;
    if (request != null && request.feedlearn === 'shownquizzeschanged') {
      fbname = request.fbname;
      fburl = request.fburl;
      if (fbname != null && fbname.length > 0) {
        root.fbname = fbname;
      }
      if (fburl != null && fburl.length > 0) {
        root.fburl = fburl;
      }
      getCookie(function(cookie){
        return addlogfb({
          type: 'shownquizzeschanged',
          'visibleids': request.visibleids,
          'shownids': request.shownids,
          'hiddenids': request.hiddenids,
          'showntimes': request.showntimes,
          fbname: fbname,
          fburl: fburl
        }, cookie);
      });
    }
    if (request != null && request.feedlearn === 'missingformat') {
      fbname = request.fbname;
      fburl = request.fburl;
      if (fbname != null && fbname.length > 0) {
        root.fbname = fbname;
      }
      if (fburl != null && fburl.length > 0) {
        root.fburl = fburl;
      }
      getCookie(function(cookie){
        var fullname;
        fullname = cookie.fullname;
        if (fullname == null || fullname === 'Anonymous User' || fullname.length === 0) {
          cookie.fullname = fbname;
        }
        if (!root.sentmissingformat) {
          root.sentmissingformat = true;
          return addlogfb({
            type: 'missingformat',
            fbname: fbname,
            fburl: fburl
          }, cookie);
        }
      });
    }
    if (request != null && request.feedlearn === 'fbstillopen') {
      fbname = request.fbname;
      fburl = request.fburl;
      if (fbname != null && fbname.length > 0) {
        root.fbname = fbname;
      }
      if (fburl != null && fburl.length > 0) {
        root.fburl = fburl;
      }
      getCookie(function(cookie){
        return addlogfb({
          type: 'fbstillopen',
          mostrecentmousemove: request.mostrecentmousemove,
          timeopened: request.timeopened,
          timesincemousemove: request.timesincemousemove,
          'visiblequizids': request.visiblequizids,
          fbname: fbname,
          fburl: fburl
        }, cookie);
      });
    }
    if (request != null && request.feedlearn === 'getformat') {
      fbname = request.fbname;
      fburl = request.fburl;
      if (fbname != null && fbname.length > 0) {
        root.fbname = fbname;
      }
      if (fburl != null && fburl.length > 0) {
        root.fburl = fburl;
      }
      return getCookie(function(cookie){
        var fullname;
        fullname = cookie.fullname;
        if (fullname == null || fullname === 'Anonymous User' || fullname.length === 0) {
          fullname = request.fbname;
          cookie.fullname = fullname;
          if (!root.sentmissingcookie) {
            root.sentmissingcookie = true;
            addlogfb({
              type: 'missingcookie',
              fbname: fbname,
              fburl: fburl
            }, cookie);
          }
        }
        return getRemoteCookies(fullname, function(remotecookie){
          var k, v, format;
          for (k in remotecookie) {
            v = remotecookie[k];
            cookie[k] = v;
          }
          format = cookie.format;
          sendResponse({
            feedlearn: true,
            format: format
          });
          chrome.tabs.query({}, function(tabs){
            var i$, len$, tab, results$ = [];
            for (i$ = 0, len$ = tabs.length; i$ < len$; ++i$) {
              tab = tabs[i$];
              results$.push(chrome.tabs.sendMessage(tab.id, {
                feedlearn: true,
                format: cookie.format
              }));
            }
            return results$;
          });
          addlog({
            type: 'fbvisit',
            fbname: fbname,
            fburl: fburl
          }, cookie);
          return addlogfb({
            type: 'fbvisit',
            fbname: fbname,
            fburl: fburl
          }, cookie);
        });
      });
    }
  });
}).call(this);
