import { useEffect, useRef, useState, useCallback } from 'react';

type WebSocketEvent = {
  type: string;
  payload: any;
};

export const useAppWebSocket = (onMessage: (event: WebSocketEvent) => void) => {
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectCountRef = useRef(0);

  const connect = useCallback(() => {
    // Determine the WS URL from the API URL. Defaults to the origin the SPA
    // was served from, mirroring the REST client's relative '/api/v1' base, so
    // a build works on any host; VITE_API_URL still overrides it when the API
    // lives somewhere else.
    const apiUrl = import.meta.env.VITE_API_URL || `${window.location.origin}/api/v1`;
    const wsUrl = apiUrl.replace(/^http/, 'ws') + '/ws';

    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        reconnectCountRef.current = 0; // Reset reconnect count on successful connection
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data) as WebSocketEvent;
          onMessage(data);
        } catch (error) {
          console.error('[WebSocket] Error parsing message:', error);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        scheduleReconnect();
      };

      ws.onerror = (error) => {
        console.error('[WebSocket] Error:', error);
        ws.close();
      };
    } catch (error) {
      console.error('[WebSocket] Connection failed:', error);
      scheduleReconnect();
    }
  }, [onMessage]);

  const scheduleReconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
    }
    
    // Exponential backoff: 1s, 2s, 4s, 8s, max 30s
    const delay = Math.min(1000 * Math.pow(2, reconnectCountRef.current), 30000);
    reconnectCountRef.current += 1;
    
    reconnectTimeoutRef.current = setTimeout(() => {
        connect();
    }, delay);
  }, [connect]);

  useEffect(() => {
    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        // Prevent onclose from triggering reconnect during unmount cleanup
        wsRef.current.onclose = null;
        wsRef.current.close();
      }
    };
  }, [connect]);

  return { isConnected };
};
