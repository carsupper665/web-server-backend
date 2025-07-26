// common/utils.go

package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

func GetRandomString(length int) string {
	key := make([]byte, length)
	for i := 0; i < length; i++ {
		key[i] = keyChars[rand.Intn(len(keyChars))]
	}
	return string(key)
}

func GetTimeString() string {
	now := time.Now()
	return fmt.Sprintf("%s%d", now.Format("20060102150405"), now.UnixNano()%1e9)
}

func SendErrorToDc(msg string) error {
	// 1. Webhook URL（請替換成你的 URL）
	url := "https://discord.com/api/webhooks/1398746978392997989/tCR8awM_NhMUFsTwC9pzif6nqw381V50xqOcyOhIitGxfEyzV1p2VHST61JefwJyhiIV"

	// 2. 準備要傳送的 JSON 負載
	payload := map[string]string{
		"content":  msg,
		"username": "ServerControllerNotify",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 3. 建立 POST 請求，並設定 Content-Type 標頭
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json") // 必須設定為 application/json :contentReference[oaicite:0]{index=0}

	// 4. 使用 http.Client 發送請求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 5. 檢查狀態碼，Discord 可能回 204 No Content 或 200 OK :contentReference[oaicite:1]{index=1}
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		// 若需要，可讀取回應內文以做除錯
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook send failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Successfully sent message: %s\n", msg)
	return nil
}
