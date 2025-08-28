from flask_sqlalchemy import SQLAlchemy
from werkzeug.security import generate_password_hash, check_password_hash
from flask_login import UserMixin
import re
import secrets
import string

db = SQLAlchemy()

def generate_secure_password(length=24):
    """Генерация безопасной случайной строки"""
    # Используем буквы (верхний и нижний регистр), цифры и специальные символы
    alphabet = string.ascii_letters + string.digits + string.punctuation
    # Убедимся, что пароль содержит хотя бы один символ каждого типа
    while True:
        password = ''.join(secrets.choice(alphabet) for _ in range(length))
        if (any(c.islower() for c in password) and 
            any(c.isupper() for c in password) and 
            any(c.isdigit() for c in password) and
            any(c in string.punctuation for c in password)):
            return password

class User(UserMixin, db.Model):
    __tablename__ = 'users'
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(80), unique=True, nullable=False)
    email = db.Column(db.String(120), unique=True, nullable=False)
    password_hash = db.Column(db.String(120), nullable=False)
    role = db.Column(db.String(20), default='user')
    
    groups = db.relationship('UserGroup', backref='user', lazy=True, cascade='all, delete-orphan')
    
    def set_password(self, password):
        self.password_hash = generate_password_hash(password)
    
    def check_password(self, password):
        return check_password_hash(self.password_hash, password)
    
    def has_group(self):
        return len(self.groups) > 0
    
    def get_group(self):
        if self.groups:
            return self.groups[0].group
        return None

class Group(db.Model):
    __tablename__ = 'groups'
    id = db.Column(db.Integer, primary_key=True)
    name = db.Column(db.String(100), unique=True, nullable=False)
    description = db.Column(db.Text)
    password_phrase = db.Column(db.String(255), nullable=True)  # Новое поле для парольной фразы
    
    users = db.relationship('UserGroup', backref='group', lazy=True, cascade='all, delete-orphan')
    
    def generate_password_phrase(self):
        """Генерирует парольную фразу для группы"""
        import random
        import string
        new_password = ''.join(random.choices(string.ascii_letters + string.digits, k=12))
        self.password_phrase = new_password
        return new_password
    
    def check_password_phrase(self, phrase):
        """Проверка парольной фразы"""
        return self.password_phrase == phrase

class UserGroup(db.Model):
    __tablename__ = 'user_groups'
    id = db.Column(db.Integer, primary_key=True)
    user_id = db.Column(db.Integer, db.ForeignKey('users.id'), nullable=False)
    group_id = db.Column(db.Integer, db.ForeignKey('groups.id'), nullable=False)
    
    __table_args__ = (db.UniqueConstraint('user_id', name='_user_unique_group'),)