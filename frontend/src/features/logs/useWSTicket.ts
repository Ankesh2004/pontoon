import { useMutation } from '@tanstack/react-query';
import { authApi } from '../../api/endpoints';
import type { WSTicket } from '../../types';

export function useWSTicket() {
  return useMutation<WSTicket>({
    mutationFn: authApi.wsTicket,
  });
}
