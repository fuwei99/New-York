package handler

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"

    "github.com/google/uuid"
)

// 在这里填写你的cookie
const COOKIE = "_ga=GA1.1.406121063.1739542075; _u_p="!w9NhM0mvCuKRv4o3xlXxQ3z5p+eaAkpv0I+xQmRZ6Fc=?eyJ1aWQiOiA1NzQ3NzAwfQ=="; _a_p="!btjZwmjgZUOOVxYJBrQldxjsPRV9EcIjA061+2LxcF4=?eyJ1aWQiOiAtMX0="; _s_p="!a5Xz4aaIF2Vfb6rFhCGQdnSeC4tRtMBsnmzszQ1Z3tc=?eyJhbm9uIjogImY2MTA2MTZhLWY5YzMtNGNlZC1hMDFiLWY2MTEzMThjNTU1ZCIsICJzZXNzaW9uX3RpbWVzdGFtcCI6IDE3Mzk2MTE4MDAuMCwgInBhc3N3b3JkX3ZlcnNpb24iOiAyLCAic3NvX2luZm8iOiB7fSwgInNlc3Npb25faWQiOiAxMjI5NTk0MH0="; _ss_p="!uT0qOP/eCyszle2OWZQQcGVM925Ig+vEgM5crODtoxs=?e30="; _ga_NM4YZNYB7G=GS1.1.1739611404.5.1.1739611826.60.0.0"

// 简化后的模型列表
var MODEL_LIST = []string{
    "ROUTE_LLM",
    "OPENAI_GPT4O",
    "CLAUDE_V3_5_SONNET",
    "OPENAI_O3_MINI",
    "OPENAI_O3_MINI_HIGH",
    "OPENAI_O1_MINI",
    "OPENAI_O1",
    "DEEPSEEK_R1_FAST",
    "DEEPSEEK_R1",
    "GEMINI_2_FLASH_THINKING",
    "GEMINI_2_FLASH",
    "GEMINI_1_5_PRO",
    "OPENAI_GPT4O_MINI",
    "XAI_GROK",
    "DEEPSEEK_V3",
    "ABACUS_SMAUG3",
    "LLAMA3_1_405B",
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type CreateConversationRequest struct {
    DeploymentId          string `json:"deploymentId"`
    Name                  string `json:"name"`
    ExternalApplicationId string `json:"externalApplicationId"`
}

type CreateConversationResponse struct {
    Success bool `json:"success"`
    Result  struct {
        DeploymentConversationId string `json:"deploymentConversationId"`
        ExternalApplicationId    string `json:"externalApplicationId"`
    } `json:"result"`
}

type ChatRequest struct {
    RequestId                string     `json:"requestId"`
    DeploymentConversationId string     `json:"deploymentConversationId"`
    Message                  string     `json:"message"`
    IsDesktop               bool       `json:"isDesktop"`
    ChatConfig              ChatConfig `json:"chatConfig"`
    LlmName                 string     `json:"llmName"`
    ExternalApplicationId   string     `json:"externalApplicationId"`
}

type ChatConfig struct {
    Timezone string `json:"timezone"`
    Language string `json:"language"`
}

type OpenAIStreamResponse struct {
    Id      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Delta struct {
            Content string `json:"content"`
        } `json:"delta"`
        Index        int    `json:"index"`
        FinishReason string `json:"finish_reason,omitempty"`
    } `json:"choices"`
}

type OpenAIResponse struct {
    Id      string `json:"id"`
    Object  string `json:"object"`
    Created int64  `json:"created"`
    Model   string `json:"model"`
    Choices []struct {
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
        FinishReason string `json:"finish_reason"`
    } `json:"choices"`
}

type AbacusResponse struct {
    Type              string  `json:"type"`
    Temp              bool    `json:"temp"`
    IsSpinny          bool    `json:"isSpinny"`
    Segment           string  `json:"segment"`
    Title             string  `json:"title"`
    IsGeneratingImage bool    `json:"isGeneratingImage"`
    MessageId         string  `json:"messageId"`
    Counter           int     `json:"counter"`
    Message_id        string  `json:"message_id"`
    Token             *string `json:"token,omitempty"`
    End               bool    `json:"end,omitempty"`
    Success           bool    `json:"success,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
    // 处理模型列表请求
    if r.URL.Path == "/v1/models" {
        handleModels(w)
        return
    }

    // 处理聊天请求
    if r.URL.Path == "/v1/chat/completions" {
        handleChat(w, r)
        return
    }

    // 默认路由
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":  "Abacus2Api Service Running...",
        "message": "MoLoveSze...",
    })
}

func handleModels(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "application/json")
    
    now := time.Now().Unix()
    var modelData []struct {
        ID         string    `json:"id"`
        Object     string    `json:"object"`
        Created    int64     `json:"created"`
        OwnedBy    string    `json:"owned_by"`
        Permission []string  `json:"permission"`
        Root       string    `json:"root"`
        Parent     *string   `json:"parent"`
    }

    for _, modelID := range MODEL_LIST {
        modelData = append(modelData, struct {
            ID         string    `json:"id"`
            Object     string    `json:"object"`
            Created    int64     `json:"created"`
            OwnedBy    string    `json:"owned_by"`
            Permission []string  `json:"permission"`
            Root       string    `json:"root"`
            Parent     *string   `json:"parent"`
        }{
            ID:         modelID,
            Object:     "model",
            Created:    now,
            OwnedBy:    "abacus",
            Permission: []string{},
            Root:       modelID,
            Parent:     nil,
        })
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "object": "list",
        "data":   modelData,
    })
}

func handleChat(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
        return
    }

    var req struct {
        Messages []Message `json:"messages"`
        Model    string    `json:"model"`
        Stream   bool      `json:"stream"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "请求格式错误", http.StatusBadRequest)
        return
    }

    // 创建会话
    convResp, err := createConversation()
    if err != nil {
        http.Error(w, "创建会话失败", http.StatusInternalServerError)
        return
    }

    // 构建消息内容
    message := buildMessage(req.Messages)

    chatReq := ChatRequest{
        RequestId:                uuid.New().String(),
        DeploymentConversationId: convResp.Result.DeploymentConversationId,
        Message:                  message,
        IsDesktop:               true,
        ChatConfig: ChatConfig{
            Timezone: "Asia/Hong_Kong",
            Language: "zh-CN",
        },
        LlmName:               req.Model,
        ExternalApplicationId: convResp.Result.ExternalApplicationId,
    }

    if req.Stream {
        handleStreamResponse(w, chatReq)
    } else {
        handleNonStreamResponse(w, chatReq)
    }
}

func createConversation() (*CreateConversationResponse, error) {
    reqBody := CreateConversationRequest{
        DeploymentId:          "14b2a314cc",
        Name:                  "New Chat",
        ExternalApplicationId: "67f8e58c3",
    }

    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", "https://pa002.abacus.ai/cluster-proxy/api/createDeploymentConversation", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    setHeaders(req)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result CreateConversationResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

func handleStreamResponse(w http.ResponseWriter, chatReq ChatRequest) error {
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", "https://pa002.abacus.ai/api/_chatLLMSendMessageSSE", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    setHeaders(req)
    req.Header.Set("Accept", "text/event-stream")
    req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    reader := bufio.NewReader(resp.Body)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                fmt.Fprintf(w, "data: [DONE]\n\n")
                return nil
            }
            return err
        }

        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        var abacusResp AbacusResponse
        if err := json.Unmarshal([]byte(line), &abacusResp); err != nil {
            continue
        }

        if abacusResp.Type == "text" && abacusResp.Title != "Thinking..." {
            streamResp := OpenAIStreamResponse{
                Id:      uuid.New().String(),
                Object:  "chat.completion.chunk",
                Created: time.Now().Unix(),
                Model:   chatReq.LlmName,
                Choices: []struct {
                    Delta struct {
                        Content string `json:"content"`
                    } `json:"delta"`
                    Index        int    `json:"index"`
                    FinishReason string `json:"finish_reason,omitempty"`
                }{
                    {
                        Delta: struct {
                            Content string `json:"content"`
                        }{
                            Content: abacusResp.Segment,
                        },
                        Index: 0,
                    },
                },
            }

            jsonResp, err := json.Marshal(streamResp)
            if err != nil {
                return err
            }

            fmt.Fprintf(w, "data: %s\n\n", jsonResp)
        }

        if abacusResp.End {
            endResp := OpenAIStreamResponse{
                Id:      uuid.New().String(),
                Object:  "chat.completion.chunk",
                Created: time.Now().Unix(),
                Model:   chatReq.LlmName,
                Choices: []struct {
                    Delta struct {
                        Content string `json:"content"`
                    } `json:"delta"`
                    Index        int    `json:"index"`
                    FinishReason string `json:"finish_reason,omitempty"`
                }{
                    {
                        Delta: struct {
                            Content string `json:"content"`
                        }{},
                        Index:        0,
                        FinishReason: "stop",
                    },
                },
            }
            jsonResp, _ := json.Marshal(endResp)
            fmt.Fprintf(w, "data: %s\n\n", jsonResp)
            fmt.Fprintf(w, "data: [DONE]\n\n")
            return nil
        }
    }
}

func handleNonStreamResponse(w http.ResponseWriter, chatReq ChatRequest) error {
    w.Header().Set("Content-Type", "application/json")

    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("POST", "https://pa002.abacus.ai/api/_chatLLMSendMessageSSE", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    setHeaders(req)
    req.Header.Set("Accept", "text/event-stream")
    req.Header.Set("Content-Type", "text/plain;charset=UTF-8")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    reader := bufio.NewReader(resp.Body)
    var content strings.Builder
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                break
            }
            return err
        }

        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }

        var abacusResp AbacusResponse
        if err := json.Unmarshal([]byte(line), &abacusResp); err != nil {
            continue
        }

        if abacusResp.Type == "text" && abacusResp.Title != "Thinking..." {
            content.WriteString(abacusResp.Segment)
        }

        if abacusResp.End {
            break
        }
    }

    openAIResp := OpenAIResponse{
        Id:      uuid.New().String(),
        Object:  "chat.completion",
        Created: time.Now().Unix(),
        Model:   chatReq.LlmName,
        Choices: []struct {
            Message struct {
                Role    string `json:"role"`
                Content string `json:"content"`
            } `json:"message"`
            FinishReason string `json:"finish_reason"`
        }{
            {
                Message: struct {
                    Role    string `json:"role"`
                    Content string `json:"content"`
                }{
                    Role:    "assistant",
                    Content: content.String(),
                },
                FinishReason: "stop",
            },
        },
    }

    return json.NewEncoder(w).Encode(openAIResp)
}

func buildMessage(messages []Message) string {
    if len(messages) == 0 {
        return ""
    }

    message := messages[len(messages)-1].Content
    var systemPrompt string
    var contextMessages []Message

    for _, msg := range messages[:len(messages)-1] {
        if msg.Role == "system" {
            systemPrompt = msg.Content
        } else {
            contextMessages = append(contextMessages, msg)
        }
    }

    fullMessage := message
    if systemPrompt != "" {
        fullMessage = fmt.Sprintf("System: %s\n\n%s", systemPrompt, message)
    }
    if len(contextMessages) > 0 {
        contextStr := ""
        for _, ctx := range contextMessages {
            contextStr += fmt.Sprintf("%s: %s\n", ctx.Role, ctx.Content)
        }
        fullMessage = fmt.Sprintf("Previous conversation:\n%s\nCurrent message: %s", contextStr, message)
    }

    return fullMessage
}

func setHeaders(req *http.Request) {
    req.Header.Set("sec-ch-ua-platform", "Windows")
    req.Header.Set("sec-ch-ua", "\"Not(A:Brand\";v=\"99\", \"Microsoft Edge\";v=\"133\", \"Chromium\";v=\"133\"")
    req.Header.Set("sec-ch-ua-mobile", "?0")
    req.Header.Set("X-Abacus-Org-Host", "apps")
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36 Edg/133.0.0.0")
    req.Header.Set("Sec-Fetch-Site", "same-site")
    req.Header.Set("Sec-Fetch-Mode", "cors")
    req.Header.Set("Sec-Fetch-Dest", "empty")
    req.Header.Set("host", "pa002.abacus.ai")
    req.Header.Set("Cookie", COOKIE)  // 使用全局定义的COOKIE常量
}
