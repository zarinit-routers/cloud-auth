import grpc
import auth_pb2
import auth_pb2_grpc
from tenacity import retry, stop_after_attempt, wait_fixed, retry_if_exception_type

class AuthClient:
    def __init__(self, host='localhost', port=50051):
        self.host = host
        self.port = port
        self.channel = None
        self.stub = None
        self._connect()
    
    def _connect(self):
        """Установка соединения с gRPC сервером"""
        try:
            self.channel = grpc.insecure_channel(f'{self.host}:{self.port}')
            self.stub = auth_pb2_grpc.AuthServiceStub(self.channel)
            # Проверяем соединение
            grpc.channel_ready_future(self.channel).result(timeout=5)
            print(f"Connected to gRPC server at {self.host}:{self.port}")
        except grpc.FutureTimeoutError:
            raise ConnectionError(f"Could not connect to gRPC server at {self.host}:{self.port}")
    
    @retry(
        stop=stop_after_attempt(3),
        wait=wait_fixed(2),
        retry=retry_if_exception_type((grpc.RpcError, ConnectionError))
    )
    def check_group(self, group_name, password_phrase=""):
        """
        Проверяет наличие группы и валидность пароля
        
        Args:
            group_name: название группы
            password_phrase: парольная фраза (опционально)
        
        Returns:
            dict: результат проверки
        """
        try:
            request = auth_pb2.GroupRequest(
                group_name=group_name,
                password_phrase=password_phrase
            )
            response = self.stub.CheckGroup(request)
            
            return {
                'exists': response.exists,
                'valid_password': response.valid_password,
                'group_description': response.group_description,
                'message': response.message,
                'error': None
            }
            
        except grpc.RpcError as e:
            error_msg = f'gRPC error: {e.details()} (code: {e.code()})'
            print(error_msg)
            return {
                'exists': False,
                'valid_password': False,
                'group_description': "",
                'message': "",
                'error': error_msg
            }
        except Exception as e:
            error_msg = f'General error: {str(e)}'
            print(error_msg)
            return {
                'exists': False,
                'valid_password': False,
                'group_description': "",
                'message': "",
                'error': error_msg
            }
    
    @retry(
        stop=stop_after_attempt(3),
        wait=wait_fixed(2),
        retry=retry_if_exception_type((grpc.RpcError, ConnectionError))
    )
    def generate_group_password(self, group_name):
        """
        Генерирует пароль для группы
        
        Args:
            group_name: название группы
        
        Returns:
            dict: результат генерации
        """
        try:
            request = auth_pb2.GenerateRequest(group_name=group_name)
            response = self.stub.GenerateGroupPassword(request)
            
            return {
                'success': response.success,
                'password': response.password,
                'error': response.error
            }
            
        except grpc.RpcError as e:
            error_msg = f'gRPC error: {e.details()} (code: {e.code()})'
            print(error_msg)
            return {
                'success': False,
                'password': '',
                'error': error_msg
            }
        except Exception as e:
            error_msg = f'General error: {str(e)}'
            print(error_msg)
            return {
                'success': False,
                'password': '',
                'error': error_msg
            }
    
    def close(self):
        """Закрывает соединение"""
        if self.channel:
            self.channel.close()
            print("gRPC connection closed")

# Глобальный экземпляр клиента
_auth_client_instance = None

def get_auth_client():
    """Получение глобального экземпляра клиента"""
    global _auth_client_instance
    if _auth_client_instance is None:
        _auth_client_instance = AuthClient()
    return _auth_client_instance

# Упрощенные функции для использования
def check_group_credentials(group_name, password=""):
    """
    Упрощенная функция для проверки группы и пароля
    """
    client = AuthClient()
    try:
        result = client.check_group(group_name, password)
        
        if result.get('error'):
            print(f"Error: {result['error']}")
            return False
        
        if not result['exists']:
            print("Group does not exist")
            return False
        
        if not result['valid_password']:
            print("Invalid password")
            return False
        
        print(f"Access granted to group: {result['group_description']}")
        return True
    finally:
        client.close()

def generate_group_password(group_name):
    """
    Упрощенная функция для генерации пароля группы
    """
    client = AuthClient()
    try:
        result = client.generate_group_password(group_name)
        
        if result['success']:
            print(f"Generated password: {result['password']}")
            return result['password']
        else:
            print(f"Error: {result['error']}")
            return None
    finally:
        client.close()

# Пример использования в коннекторе
if __name__ == '__main__':
    # Пример проверки группы
    print("Testing group check:")
    result = check_group_credentials("test_group", "password123")
    print(f"Access granted: {result}")
    
    # Пример генерации пароля
    print("\nTesting password generation:")
    password = generate_group_password("test_group")
    if password:
        print(f"New password: {password}")