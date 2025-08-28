# run_services.py
import subprocess
import sys
import time
import threading
import os
import signal

# Глобальные переменные для процессов
processes = []

def signal_handler(sig, frame):
    """Обработчик сигналов для graceful shutdown"""
    print("\nShutting down services...")
    for process in processes:
        try:
            process.terminate()
        except:
            pass
    sys.exit(0)

def run_service(command, name):
    """Запуск сервиса с логированием"""
    try:
        print(f"Starting {name}...")
        process = subprocess.Popen(
            command,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
            universal_newlines=True
        )
        processes.append(process)
        
        # Логирование вывода в реальном времени
        def log_output():
            while True:
                output = process.stdout.readline()
                if output == '' and process.poll() is not None:
                    break
                if output:
                    print(f"[{name}] {output.strip()}")
            process.stdout.close()
        
        log_thread = threading.Thread(target=log_output, daemon=True)
        log_thread.start()
        
        return process
        
    except Exception as e:
        print(f"Error starting {name}: {e}")
        return None

def main():
    """Основная функция запуска"""
    print("Starting both Flask app and gRPC server...")
    
    # Регистрируем обработчики сигналов
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    # Запускаем gRPC сервер
    grpc_process = run_service([sys.executable, "grpc_server.py"], "gRPC")
    
    if not grpc_process:
        print("Failed to start gRPC server")
        return
    
    # Ждем немного, чтобы gRPC сервер успел запуститься
    time.sleep(3)
    
    # Запускаем Flask приложение
    flask_process = run_service([sys.executable, "app.py"], "Flask")
    
    if not flask_process:
        print("Failed to start Flask app")
        grpc_process.terminate()
        return
    
    # Мониторим процессы
    try:
        while True:
            # Проверяем, работают ли процессы
            if grpc_process.poll() is not None:
                print("gRPC server stopped unexpectedly")
                flask_process.terminate()
                break
                
            if flask_process.poll() is not None:
                print("Flask app stopped unexpectedly")
                grpc_process.terminate()
                break
                
            time.sleep(1)
            
    except KeyboardInterrupt:
        signal_handler(None, None)

if __name__ == '__main__':
    main()