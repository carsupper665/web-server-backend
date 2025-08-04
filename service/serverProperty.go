//service/setProperty.go

package service

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func read(dir string) (*os.File, error) {
	path := dir + "/server.properties"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		header := fmt.Sprintf("#Minecraft server properties\n#Generated on %s\n", time.Now().Format(time.RFC1123))
		if err := os.WriteFile(path, []byte(header), 0644); err != nil {
			return nil, err
		}
	}

	f, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	return f, nil
}

func backUp(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func UpdateProperty(workDir, key, value string) error {
	path := workDir + "/server.properties"
	_ = backUp(path, path+".bak")

	f, err := read(path)

	if err != nil {
		return err
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// 保留註解/空行原樣
		if strings.HasPrefix(trimmedLine, "#") || trimmedLine == "" {
			lines = append(lines, line)
			continue
		}

		parts := strings.SplitN(trimmedLine, "=", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
			// 改成新的值（覆寫）
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
			found = true
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	tmpPath := path + ".tmp"
	joined := strings.Join(lines, "\n")
	if !strings.HasSuffix(joined, "\n") {
		joined += "\n" // 保留 trailing newline
	}
	if err := os.WriteFile(tmpPath, []byte(joined), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	return nil

}

func GetPropertyText(workDir string) (string, error) {
	f, err := read(workDir)

	if err != nil {
		return "", nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(lines, "\n"), nil
}

func ReplaceProperty(workDir string, texts string) error {

	path := workDir + "/server.properties"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		header := fmt.Sprintf("#Minecraft server properties\n#Generated on %s\n", time.Now().Format(time.RFC1123))
		if err := os.WriteFile(path, []byte(header), 0644); err != nil {
			return err
		}
	}

	_ = backUp(path, path+".bak")
	tmpPath := path + ".tmp"
	if !strings.HasSuffix(texts, "\n") {
		texts += "\n" // 保留 trailing newline
	}
	if err := os.WriteFile(tmpPath, []byte(texts), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	return nil
}
