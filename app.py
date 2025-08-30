from flask import Flask, jsonify, render_template, request, redirect, url_for, flash
from flask_login import LoginManager, login_user, login_required, logout_user, current_user
from models import db, User, Group, UserGroup
from functools import wraps
import re
import os
from flask_migrate import Migrate 

app = Flask(__name__, static_folder='static')
app.config['SECRET_KEY'] = 'your-secret-key'
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://postgres:1@localhost/adminka'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

db.init_app(app)
migrate = Migrate(app, db)
login_manager = LoginManager()
login_manager.init_app(app)
login_manager.login_view = 'login'

def is_valid_email(email):
    pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
    return re.match(pattern, email) is not None

@login_manager.user_loader
def load_user(user_id):
    return User.query.get(int(user_id))

def admin_required(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        if not current_user.is_authenticated or current_user.role != 'admin':
            flash('Требуются права администратора')
            return redirect(url_for('dashboard'))
        return f(*args, **kwargs)
    return decorated_function

@app.route('/')
def index():
    return redirect(url_for('login'))

@app.route('/auth/login', methods=['GET', 'POST'])
def login():
    if current_user.is_authenticated:
        return redirect(url_for('dashboard'))
    
    if request.method == 'POST':
        email = request.form['email']
        password = request.form['password']
        user = User.query.filter_by(email=email).first()
        
        if user and user.check_password(password):
            login_user(user)
            next_page = request.args.get('next')
            return redirect(next_page or url_for('dashboard'))
        else:
            flash('Неверные учетные данные')
    
    return render_template('login.html')

@app.route('/auth/logout')
@login_required
def logout():
    logout_user()
    return redirect(url_for('login'))

@app.route('/auth/dashboard')
@login_required
def dashboard():
    return render_template('dashboard.html', user=current_user)

@app.route('/auth/profile', methods=['GET', 'POST'])
@login_required
def profile():
    groups = Group.query.all()
    
    if request.method == 'POST':
        if 'update_profile' in request.form:
            new_username = request.form['username']
            new_email = request.form['email']
            
            # Проверяем, не занят ли новый username другим пользователем
            if new_username != current_user.username and User.query.filter_by(username=new_username).first():
                flash('Это имя пользователя уже занято')
                return redirect(url_for('profile'))
            
            # Проверяем, не занят ли новый email другим пользователем
            if new_email != current_user.email and User.query.filter_by(email=new_email).first():
                flash('Этот email уже занят')
                return redirect(url_for('profile'))
            
            # Проверяем валидность email
            if not is_valid_email(new_email):
                flash('Неверный формат email')
                return redirect(url_for('profile'))
            
            current_user.username = new_username
            current_user.email = new_email
            
            if request.form['password']:
                current_user.set_password(request.form['password'])
            
            db.session.commit()
            flash('Профиль обновлен')
        
        elif 'assign_group' in request.form and current_user.role == 'admin':
            group_id = request.form['group_id']
            user_id = request.form['user_id']
            
            user = User.query.get(user_id)
            group = Group.query.get(group_id)
            
            if user and group:
                # Удаляем существующие группы пользователя
                UserGroup.query.filter_by(user_id=user_id).delete()
                
                # Добавляем новую группу
                user_group = UserGroup(user_id=user_id, group_id=group_id)
                db.session.add(user_group)
                db.session.commit()
                flash(f'Пользователю {user.username} назначена группа {group.name}')
    
    return render_template('profile.html', groups=groups)

# Админские роуты
@app.route('/auth/admin/users')
@login_required
@admin_required
def admin_users():
    users = User.query.all()
    return render_template('admin/users.html', users=users)

@app.route('/auth/admin/create-user', methods=['GET', 'POST'])
@login_required
@admin_required
def create_user():
    if request.method == 'POST':
        username = request.form['username']
        email = request.form['email']
        password = request.form['password']
        role = request.form['role']
        
        if User.query.filter_by(username=username).first():
            flash('Пользователь с таким именем уже существует')
            return redirect(url_for('create_user'))
        
        if User.query.filter_by(email=email).first():
            flash('Пользователь с таким email уже существует')
            return redirect(url_for('create_user'))
        
        if not is_valid_email(email):
            flash('Неверный формат email')
            return redirect(url_for('create_user'))
        
        user = User(username=username, email=email, role=role)
        user.set_password(password)
        db.session.add(user)
        db.session.commit()
        flash('Пользователь создан')
        return redirect(url_for('admin_users'))
    
    return render_template('admin/create_user.html')

@app.route('/auth/admin/edit-user/<int:user_id>', methods=['GET', 'POST'])
@login_required
@admin_required
def edit_user(user_id):
    user = User.query.get_or_404(user_id)
    groups = Group.query.all()
    
    if request.method == 'POST':
        username = request.form['username']
        email = request.form['email']
        role = request.form['role']
        
        # Проверяем, не занят ли username другим пользователем
        if username != user.username and User.query.filter_by(username=username).first():
            flash('Это имя пользователя уже занято')
            return redirect(url_for('edit_user', user_id=user_id))
        
        # Проверяем, не занят ли email другим пользователем
        if email != user.email and User.query.filter_by(email=email).first():
            flash('Этот email уже занят')
            return redirect(url_for('edit_user', user_id=user_id))
        
        if not is_valid_email(email):
            flash('Неверный формат email')
            return redirect(url_for('edit_user', user_id=user_id))
        
        user.username = username
        user.email = email
        user.role = role
        
        if request.form['password']:
            user.set_password(request.form['password'])
        
        db.session.commit()
        flash('Пользователь обновлен')
        return redirect(url_for('admin_users'))
    
    return render_template('admin/edit_user.html', user=user, groups=groups)

@app.route('/auth/admin/delete-user/<int:user_id>')
@login_required
@admin_required
def delete_user(user_id):
    if user_id == current_user.id:
        flash('Нельзя удалить самого себя')
        return redirect(url_for('admin_users'))
    
    user = User.query.get_or_404(user_id)
    db.session.delete(user)
    db.session.commit()
    flash('Пользователь удален')
    return redirect(url_for('admin_users'))

@app.route('/auth/admin/groups')
@login_required
@admin_required
def admin_groups():
    groups = Group.query.all()
    users = User.query.all()
    return render_template('admin/groups.html', groups=groups, users=users)

@app.route('/auth/admin/create-group', methods=['GET', 'POST'])
@login_required
@admin_required
def create_group():
    if request.method == 'POST':
        name = request.form['name']
        description = request.form['description']
        
        if Group.query.filter_by(name=name).first():
            flash('Группа уже существует')
            return redirect(url_for('create_group'))
        
        group = Group(name=name, description=description)
        db.session.add(group)
        db.session.commit()
        flash('Группа создана')
        return redirect(url_for('admin_groups'))
    
    return render_template('admin/create_group.html')

@app.route('/auth/admin/edit-group/<int:group_id>', methods=['GET', 'POST'])
@login_required
@admin_required
def edit_group(group_id):
    group = Group.query.get_or_404(group_id)
    
    if request.method == 'POST':
        name = request.form['name']
        description = request.form['description']
        
        # Проверяем, не занято ли имя другой группой
        if name != group.name and Group.query.filter_by(name=name).first():
            flash('Группа с таким именем уже существует')
            return redirect(url_for('edit_group', group_id=group_id))
        
        group.name = name
        group.description = description
        db.session.commit()
        flash('Группа обновлена')
        return redirect(url_for('admin_groups'))
    
    return render_template('admin/edit_group.html', group=group)

@app.route('/auth/admin/delete-group/<int:group_id>')
@login_required
@admin_required
def delete_group(group_id):
    group = Group.query.get_or_404(group_id)
    db.session.delete(group)
    db.session.commit()
    flash('Группа удалена')
    return redirect(url_for('admin_groups'))


@app.route('/auth/admin/generate-group-password/<int:group_id>', methods=['POST'])
@login_required
@admin_required
def generate_group_password(group_id):
    group = Group.query.get_or_404(group_id)
    new_password = group.generate_password_phrase()
    db.session.commit()
    return jsonify({'success': True, 'password': new_password})

@app.route('/auth/admin/clear-group-password/<int:group_id>', methods=['POST'])
@login_required
@admin_required
def clear_group_password(group_id):
    group = Group.query.get_or_404(group_id)
    group.password_phrase = None
    db.session.commit()
    return jsonify({'success': True})

@app.route('/auth/admin/add-user-to-group', methods=['POST'])
@login_required
@admin_required
def add_user_to_group():
    user_id = request.form['user_id']
    group_id = request.form['group_id']
    
    user = User.query.get(user_id)
    group = Group.query.get(group_id)
    
    if not user or not group:
        flash('Пользователь или группа не найдены')
        return redirect(url_for('admin_groups'))
    
    # Удаляем существующие группы пользователя (максимум 1 группа)
    UserGroup.query.filter_by(user_id=user_id).delete()
    
    # Проверяем, не состоит ли уже пользователь в группе
    if not UserGroup.query.filter_by(user_id=user_id, group_id=group_id).first():
        user_group = UserGroup(user_id=user_id, group_id=group_id)
        db.session.add(user_group)
        db.session.commit()
        flash(f'Пользователь {user.username} добавлен в группу {group.name}')
    else:
        flash('Пользователь уже состоит в этой группе')
    
    return redirect(url_for('admin_groups'))


# API endpoints
@app.route('/auth/api/check-group', methods=['POST'])
def api_check_group():
    try:
        data = request.get_json()
        group_name = data.get('group_name')
        password_phrase = data.get('password_phrase', '')
        
        if not group_name:
            return jsonify({'error': 'Group name is required'}), 400
        
        # Используем глобальный клиент или создаем новый
        from grpc_client import get_auth_client
        client = get_auth_client()
        result = client.check_group(group_name, password_phrase)
        
        return jsonify(result)
        
    except Exception as e:
        return jsonify({
            'error': f'Server error: {str(e)}',
            'exists': False,
            'valid_password': False,
            'group_description': "",
            'message': ""
        }), 500

@app.route('/auth/api/generate-password', methods=['POST'])
@login_required
@admin_required
def api_generate_password():
    try:
        data = request.get_json()
        group_name = data.get('group_name')
        
        if not group_name:
            return jsonify({'error': 'Group name is required'}), 400
        
        from grpc_client import get_auth_client
        client = get_auth_client()
        result = client.generate_group_password(group_name)
        
        return jsonify(result)
        
    except Exception as e:
        return jsonify({
            'success': False,
            'password': '',
            'error': f'Server error: {str(e)}'
        }), 500
    

    
@app.route('/auth/admin/remove-user-from-group/<int:user_group_id>')
@login_required
@admin_required
def remove_user_from_group(user_group_id):
    user_group = UserGroup.query.get_or_404(user_group_id)
    db.session.delete(user_group)
    db.session.commit()
    flash('Пользователь удален из группы')
    return redirect(url_for('admin_groups'))

ROOT_USER_NAME = 'root'
ROOT_USER_EMAIL = os.environ.get('ROOT_USER_EMAIL') or 'root@admin.com'
ROOT_USER_PASSWORD = os.environ.get('ROOT_USER_PASSWORD') or 'admin123'

APP_DEBUG = os.environ.get('FLASK_ENV') == 'development' or False

def create_admin():
    admin = User(username=ROOT_USER_NAME, email=ROOT_USER_EMAIL, role='admin')
    admin.set_password(ROOT_USER_PASSWORD)
    db.session.add(admin)
    db.session.commit()

if __name__ == '__main__':
    with app.app_context():
        db.create_all()
        # Создаем администратора по умолчанию
        if not User.query.filter_by(email=ROOT_USER_EMAIL).first():
            create_admin()

    # Запускаем Flask приложение
    app.run(debug=APP_DEBUG, host='0.0.0.0', port=5001)