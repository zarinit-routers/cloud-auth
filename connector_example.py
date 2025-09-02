from grpc_client import AuthClient

def check_group_credentials(group_name, password):
    client = AuthClient()
    result = client.check_group(group_name, password)
    client.close()
    
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

# Пример использования
if __name__ == '__main__':
    # Проверка группы
    success = check_group_credentials("test_group", "my_password")
    print(f"Access: {success}")