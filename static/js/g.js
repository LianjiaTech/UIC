function err_message_quietly(msg, f) {
	$.layer({
		title : false,
		closeBtn : false,
		time : 2,
		dialog : {
			msg : msg
		},
		end : f
	});
}

function ok_message_quietly(msg, f) {
	$.layer({
		title : false,
		closeBtn : false,
		time : 1,
		dialog : {
			msg : msg,
			type : 1
		},
		end : f
	});
}

function my_confirm(msg, btns, yes_func, no_func) {
	$.layer({
		shade : [ 0 ],
		area : [ 'auto', 'auto' ],
		dialog : {
			msg : msg,
			btns : 2,
			type : 4,
			btn : btns,
			yes : yes_func,
			no : no_func
		}
	});
}

// - business function -

function update_profile() {
	$.post('/me/profile', {
		'cnname' : $("#cnname").val(),
		'email' : $("#email").val(),
		'phone' : $("#phone").val(),
		'im' : $("#im").val(),
		'qq' : $("#qq").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly("更新成功：）");
		}
	});
}

function change_password() {
	$.post('/me/chpwd', {
		'old_password' : $("#old_password").val(),
		'new_password' : $("#new_password").val(),
		'repeat_password' : $("#repeat_password").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly("密码修改成功：）");
		}
	});
}

function register() {
	$.post('/auth/register', {
		'name' : $('#name').val(),
		'password' : $("#password").val(),
		'repeat_password' : $("#repeat_password").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly('sign up successfully', function() {
				location.href = '/auth/login';
			});
		}
	});
}

function login() {
    useLdap = '1'
	$.post('/auth/login', {
		'name' : $('#name').val(),
		'password' : $("#password").val(),
		'ldap' : useLdap,
		'sig': $("#sig").val(),
		'callback': $("#callback").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly('sign in successfully', function() {
				var redirect_url = '/me/info';
				if (json.data.length > 0) {
					redirect_url = json.data;
				}
				location.href = redirect_url;
			});
		}
	});
}

function query_user() {
	var query = $("#query").val();
	location.href = "/me/users?query=" + query;
}

function query_team() {
	var query = $("#query").val();
	location.href = "/me/teams?query=" + query;
}

function create_user() {
	$.post('/me/user/c', {
		'name' : $("#name").val(),
		'cnname' : $("#cnname").val(),
		'email' : $("#email").val(),
		'phone' : $("#phone").val(),
		'im' : $("#im").val(),
		'qq' : $("#qq").val(),
		'password' : $("#password").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly("create user successfully");
		}
	});
}

function edit_user(name) {
	$.post('/target-user/edit?name='+name, {
		'cnname' : $("#cnname").val(),
		'email' : $("#email").val(),
		'phone' : $("#phone").val(),
		'im' : $("#im").val(),
		'qq' : $("#qq").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly("更新成功：）");
		}
	});
}

function reset_password(id) {
	$.post('/target-user/chpwd?id='+id, {
		'password' : $("#password").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly("密码重置成功：）");
		}
	});
}

function create_team() {
	$.post('/me/team/c', {
		'name' : $("#name").val(),
		'resume' : $("#resume").val(),
		'admins' : $("#teamadmins").val(),
		'email' : $("#email").val(),
		'users' : $("#users").val()
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly('create team successfully');
            location.href = "/me/teams";
		}
	});
}

function edit_team(tid) {
	$.post('/target-team/edit', {
		'resume' : $("#resume").val(),
		'users' : $("#users").val(),
        'teamemail' : $("#teamemail").val(),
		'admins' : $("#teamadmins").val(),
		'id': tid
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
		} else {
			ok_message_quietly('edit team successfully');
            location.reload();
		}
	});
}

function delete_user(uid) {
	my_confirm("真的要删除么？通常只有离职的时候才需要删除", [ '确定', '取消' ], function() {
		$.post('/target-user/delete', {
			'id' : uid
		}, function(json) {
			if (json.msg.length > 0) {
				err_message_quietly(json.msg);
			} else {
				ok_message_quietly('delete user successfully', function() {
					location.reload();
				});
			}
		});
	}, function() {
	});
}

function delete_team(id) {
	my_confirm("真的真的要删除么？", [ '确定', '取消' ], function() {
		$.get('/target-team/delete?id='+id, {}, function(json) {
			if (json.msg.length > 0) {
				err_message_quietly(json.msg);
			} else {
				ok_message_quietly('delete team successfully', function() {
					location.reload();
				});
			}
		});
	}, function() {
	});
}

function set_role(name, obj) {
	var role = obj.checked ? '2' : '0';
	$.post('/target-user/role?name='+name, {
		'role' : role
	}, function(json) {
		if (json.msg.length > 0) {
			err_message_quietly(json.msg);
			location.reload();
		} else {
			if (role == '2') {
				ok_message_quietly('成功设置为管理员：）');
			} else if (role == '0') {
				ok_message_quietly('成功取消管理员权限：）');
			}
		}
	});
}

function user_detail(uid) {
	$("#user_detail_div").load("/user/detail?id=" + uid);
	$.layer({
		type : 1,
		shade : [ 0.5, '#000' ],
		shadeClose : true,
		closeBtn : [ 0, true ],
		area : [ '450px', '240px' ],
		title : false,
		border : [ 0 ],
		page : {
			dom : '#user_detail_div'
		}
	});
}
