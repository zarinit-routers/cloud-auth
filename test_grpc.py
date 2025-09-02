# test_grpc.py
from grpc_client import check_group_credentials, generate_group_password

print("Testing gRPC connection...")

# Сначала создадим тестовую группу через веб-интерфейс
# или используем существующую

# Тест проверки несуществующей группы
print("\n1. Testing non-existent group:")
result = check_group_credentials("admins", "pzLMAaOiakBn")
print(f"Result: {result}")

# Тест генерации пароля (должен вернуть ошибку для несуществующей группы)
print("\n2. Testing password generation for non-existent group:")
password = generate_group_password("admins")
print(f"Result: {password}")

print("\nTest completed! Check the logs above.")