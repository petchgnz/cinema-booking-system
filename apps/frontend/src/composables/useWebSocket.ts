import { ref, onUnmounted } from 'vue';
import type { SeatEvent } from '../types';

export function useWebSocket(showtimeId: string) {
  const isConnected = ref<boolean>(false);
  let ws: WebSocket | null = null;

  function connect(onMessage: (event: SeatEvent) => void): void {
    const url = `ws://localhost:8080/ws/showtimes/${showtimeId}`;
    ws = new WebSocket(url);

    ws.onopen = () => {
      isConnected.value = true;
      console.log(`[WS] Connected to showtime: ${showtimeId}`);
    };

    ws.onmessage = (e: MessageEvent) => {
      const event: SeatEvent = JSON.parse(e.data);
      onMessage(event);
    };

    ws.onclose = () => {
      isConnected.value = false;
      console.log(`[WS] Disconnected`);
    };
  }

  function disconnect(): void {
    ws?.close();
  }

  onUnmounted(() => {
    disconnect();
  });

  return { isConnected, connect, disconnect }
}
