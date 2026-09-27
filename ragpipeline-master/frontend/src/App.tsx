import React, { useState, useRef, useEffect } from 'react';
import { Send, Bot, User, Loader2, Sparkles, BookOpen } from 'lucide-react';

interface ChatMessage {
  id: string;
  role: 'user' | 'ai';
  content: string;
  isStreaming: boolean;
  sources?: any[];
}

// Generate a random session ID for testing
const SESSION_ID = Math.random().toString(36).substring(7);
const API_URL = import.meta.env.VITE_API_URL || '';

// Mock JWT Token that the backend requires
// The backend needs a JWT with 'grade' and 'session_id' claims
const getAuthToken = () => {
    // In a real app, you would retrieve this from auth state.
    // We construct a mock JWT (header.payload.signature)
    const payload = {
        sub: 'student-123',
        grade: 8,
        session_id: SESSION_ID,
        exp: Math.floor(Date.now() / 1000) + (60 * 60)
    };
    const b64Payload = btoa(JSON.stringify(payload));
    const b64Header = btoa(JSON.stringify({ alg: "HS256", typ: "JWT" }));
    return `${b64Header}.${b64Payload}.MockSignature`;
}

function App() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input,
      isStreaming: false
    };

    const aiMsgId = (Date.now() + 1).toString();
    const initialAiMsg: ChatMessage = {
      id: aiMsgId,
      role: 'ai',
      content: '',
      isStreaming: true
    };

    setMessages(prev => [...prev, userMsg, initialAiMsg]);
    setInput('');
    setIsLoading(true);

    try {
      const response = await fetch(`${API_URL}/query`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getAuthToken()}`
        },
        body: JSON.stringify({
          query: userMsg.content,
          session_id: SESSION_ID
        })
      });

      if (!response.ok) {
        throw new Error(`API error: ${response.statusText}`);
      }

      // Handle Server-Sent Events (SSE) stream
      const reader = response.body?.getReader();
      const decoder = new TextDecoder('utf-8');

      if (!reader) throw new Error('No readable stream');

      let currentContent = '';
      
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value, { stream: true });
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('event: complete')) {
            break;
          }
          if (line.startsWith('data: ')) {
            const dataStr = line.substring(6);
            if (dataStr === '[DONE]') break;
            
            try {
              // Try parsing as JSON in case of structured data
              const obj = JSON.parse(dataStr);
              if (obj.status === 'success') break;
              if (obj.code) { // It's an error event
                  currentContent += `\n\n**Error:** ${obj.message}`;
              }
            } catch {
              // It's raw text chunk from Gemini SSE
              currentContent += dataStr;
              
              // Only format newlines strictly if it's literal \n
              // Usually SSE tokens might have actual raw characters
              const formattedContent = currentContent.replace(/\\n/g, '\n');

              setMessages(prev => prev.map(msg => 
                msg.id === aiMsgId 
                  ? { ...msg, content: formattedContent } 
                  : msg
              ));
            }
          }
        }
      }

      setMessages(prev => prev.map(msg => 
        msg.id === aiMsgId 
          ? { ...msg, isStreaming: false } 
          : msg
      ));

    } catch (error) {
      console.error(error);
      setMessages(prev => prev.map(msg => 
        msg.id === aiMsgId 
          ? { ...msg, content: `Error connecting to RAG Backend: ${(error as Error).message}`, isStreaming: false } 
          : msg
      ));
    } finally {
      setIsLoading(false);
    }
  };

  // Helper to format citations in the text
  const formatContentWithCitations = (content: string) => {
    // Look for patterns like [Page 42]
    const parts = content.split(/(\[Page \d+\])/g);
    
    return parts.map((part, i) => {
      if (part.match(/\[Page \d+\]/)) {
        return <span key={i} className="citation">{part}</span>;
      }
      return <span key={i}>{part}</span>;
    });
  };

  return (
    <div className="app-container">
      <header className="app-header">
        <div className="brand-title">
          <BookOpen size={28} />
          Visionary Science Tutor
        </div>
        <div style={{ display: 'flex', gap: '8px', color: 'var(--text-muted)'}}>
          <Sparkles size={18} color="var(--accent-secondary)" />
          <span style={{fontSize: '0.85rem'}}>Grade 8 Active</span>
        </div>
      </header>

      <main className="chat-container">
        {messages.length === 0 && (
          <div style={{ flex: 1, display: 'flex', flexDirection: 'column', justifyContent: 'center', alignItems: 'center', opacity: 0.6 }}>
            <Bot size={64} color="var(--accent-primary)" style={{ marginBottom: '1rem' }} />
            <h2>What would you like to learn today?</h2>
            <p style={{ marginTop: '0.5rem', color: 'var(--text-muted)'}}>Ask me anything about Class 8 Science.</p>
          </div>
        )}

        {messages.map(msg => (
          <div key={msg.id} className={`message-wrapper ${msg.role}`}>
            <div className={`message-bubble ${msg.role === 'ai' ? 'ai-content' : ''}`}>
              {msg.role === 'ai' && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px', opacity: 0.8, fontSize: '0.8rem', color: 'var(--accent-primary)' }}>
                  <Bot size={16} /> Tutor
                </div>
              )}
              {msg.role === 'user' && (
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px', opacity: 0.8, fontSize: '0.8rem', justifyContent: 'flex-end' }}>
                  Student <User size={16} />
                </div>
              )}
              
              <div style={{ whiteSpace: 'pre-wrap' }}>
                {msg.role === 'ai' ? formatContentWithCitations(msg.content) : msg.content}
              </div>

              {msg.isStreaming && (
                <div className="typing-indicator" style={{ display: 'inline-flex', marginLeft: '10px' }}>
                  <span></span><span></span><span></span>
                </div>
              )}
            </div>
          </div>
        ))}
        <div ref={messagesEndRef} />
      </main>

      <form className="input-container" onSubmit={handleSubmit}>
        <input
          type="text"
          className="chat-input"
          placeholder="e.g. Explan photosynthesis or what is force..."
          value={input}
          onChange={(e) => setInput(e.target.value)}
          disabled={isLoading}
          autoFocus
        />
        <button 
          type="submit" 
          className="send-button"
          disabled={!input.trim() || isLoading}
        >
          {isLoading ? <Loader2 size={20} className="spin" /> : <Send size={20} />}
        </button>
      </form>
      
      {/* Simple global spin animation for the loader */}
      <style>{`
        .spin { animation: spin 1s linear infinite; }
        @keyframes spin { 100% { transform: rotate(360deg); } }
      `}</style>
    </div>
  );
}

export default App;
