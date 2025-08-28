import grpc
from concurrent import futures
import time
from flask import Flask
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker
import auth_pb2
import auth_pb2_grpc
import os
import sys

# Добавляем текущую директорию в путь для импорта
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# Создаем минимальное Flask приложение для контекста
app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = 'postgresql://postgres:1@localhost/adminka'
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False

# Создаем engine и сессию для базы данных
engine = create_engine(app.config['SQLALCHEMY_DATABASE_URI'])
Session = sessionmaker(bind=engine)

# Импортируем модели после настройки
from models import Group

class AuthServicer(auth_pb2_grpc.AuthServiceServicer):
    def CheckGroup(self, request, context):
        try:
            session = Session()
            group = session.query(Group).filter_by(name=request.group_name).first()
            
            if not group:
                return auth_pb2.GroupResponse(
                    exists=False,
                    valid_password=False,
                    group_description="",
                    message="Group not found"
                )
            
            valid_password = False
            message = "Group exists"
            
            if group.password_phrase:
                if request.password_phrase:
                    valid_password = (group.password_phrase == request.password_phrase)
                    message = "Password correct" if valid_password else "Invalid password"
                else:
                    message = "Password required"
            else:
                message = "Group has no password set"
                valid_password = True  # Если пароль не установлен, считаем валидным
            
            session.close()
            
            return auth_pb2.GroupResponse(
                exists=True,
                valid_password=valid_password,
                group_description=group.description or "",
                message=message
            )
            
        except Exception as e:
            print(f"Error in CheckGroup: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Server error: {str(e)}")
            return auth_pb2.GroupResponse(
                exists=False,
                valid_password=False,
                group_description="",
                message=f"Error: {str(e)}"
            )
    
    def GenerateGroupPassword(self, request, context):
        try:
            session = Session()
            group = session.query(Group).filter_by(name=request.group_name).first()
            
            if not group:
                session.close()
                return auth_pb2.GenerateResponse(
                    success=False,
                    password="",
                    error="Group not found"
                )
            
            # Генерируем пароль (простая реализация)
            import random
            import string
            new_password = ''.join(random.choices(string.ascii_letters + string.digits, k=12))
            
            group.password_phrase = new_password
            session.commit()
            session.close()
            
            return auth_pb2.GenerateResponse(
                success=True,
                password=new_password,
                error=""
            )
            
        except Exception as e:
            print(f"Error in GenerateGroupPassword: {str(e)}")
            return auth_pb2.GenerateResponse(
                success=False,
                password="",
                error=str(e)
            )

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    auth_pb2_grpc.add_AuthServiceServicer_to_server(AuthServicer(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    print("gRPC server started on port 50051")
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        server.stop(0)
        print("gRPC server stopped")

if __name__ == '__main__':
    serve()