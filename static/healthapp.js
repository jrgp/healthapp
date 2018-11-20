function init_healthapp() {
  Handlebars.registerHelper('booltostr', function(str) {
    return str ? 'Yes' : 'no';
  });

  Handlebars.registerHelper('humansize', function(size) {
    if (size < (1024 * 1024 * 1024)) {
      return (size / (1024 * 1024)).toFixed(2) + 'MB';
    }
    if (size < (1024 * 1024 * 1024 * 1024)) {
      return (size / (1024 * 1024 * 1024)).toFixed(2) + 'GB';
    }
    if (size < (1024 * 1024 * 1024 * 1024 * 1024)) {
      return (size / (1024 * 1024 * 1024 * 1024)).toFixed(2) + 'TB';
    }
  });

  Handlebars.registerHelper('usagecolor', function(size) {
    size = parseInt(size);
    if (size > 95) {
      return 'danger';
    }
    if (size > 70) {
      return 'warning';
    }
    return 'success';
  });

  function sort_fs_desc(filesystems) {
    return filesystems.sort(function(a, b) {
      return a.Pct < b.Pct;
    });
  }

  function get_servers(callback) {
    $.get('/api/v0/servers', callback);
  }

  function get_server(servername, callback) {
    $.get('/api/v0/status/' + servername, callback);
  }

  function get_alerts(callback) {
    $.get('/api/v0/alerts', callback);
  }

  function get_alert(alert_id, callback) {
    $.get('/api/v0/alert/' + alert_id, callback);
  }

  var server_list = Handlebars.compile($('#server-list-template').html()),
      alert_list = Handlebars.compile($('#alert-list-template').html()),
      flash_template = Handlebars.compile($('#flash-template').html()),
      server_view = Handlebars.compile($('#server-view-template').html()),
      alert_view = Handlebars.compile($('#alert-view-template').html()),
      user_pagination_interval = 100,
      $content = $('#content'),
      $flashes = $('#flashes'),
      $title = $('h1'),
      router = new Navigo(null, false, '#!'),
      last_flash = null;

  function flash(type, message) {
    last_flash = {type: type, message: message};
  }

  function render_page(title, contents) {
    document.title = title + ' :: HealthApp';
    $title.text(title);
    $content.html(contents);
    router.updatePageLinks();
    if (last_flash) {
      $flashes.html(flash_template(last_flash));
      last_flash = null;
    } else {
      $flashes.empty();
      last_flash = null;
    }
  };

  function servers_list_page(params) {
    get_servers(function(data) {
      for (var i = 0; i < data.servers.length; i++) {
        if (data.servers[i].info.Filesystems.length > 0) {
          sort_fs_desc(data.servers[i].info.Filesystems);
          data.servers[i].fullest_disk = data.servers[i].info.Filesystems[0];
        } else {
          data.servers[i].fullest_disk = false;
        }
      }
      render_page('Servers', server_list(data));
    });
  }

  function server_view_page(params) {
      get_server(params.servername, function(data) {
        var filesystems = data.Filesystems;
        delete data.Filesystems;
        sort_fs_desc(filesystems)
        render_page(params.servername, server_view({info: data, filesystems: filesystems}));
      });
  }

  function alerts_list_page(params) {
    get_alerts(function(data) {
      render_page('Alerts', alert_list(data));
    });
  }

  function alert_view_page(params) {
      get_alert(params.alertid, function(data) {
        render_page(data.server.name + ' ' + data.human_bad, alert_view(data));
      });
  }

  function alert_row_click(event) {
    var alert_id = $(event.target).closest('tr').data('id');
    router.navigate('/alert/' + alert_id);
  }

  function server_row_click(event) {
    var server_name = $(event.target).closest('tr').data('name');
    router.navigate('/server/' + server_name);
  }

  $content.on('click', '.alert-row', alert_row_click)
  $content.on('click', '.server-row', server_row_click)

  router.on({
    '/': servers_list_page,
    '/alerts': alerts_list_page,
    '/alert/:alertid': alert_view_page,
    '/server/:servername': server_view_page,

  }).resolve();
}
