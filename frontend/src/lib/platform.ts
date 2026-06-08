import { useQuery } from '@tanstack/react-query'
import { appApi } from '@/lib/api/app'

export function isDarwinPlatform(platform: string | undefined): boolean {
  return platform?.startsWith('darwin/') ?? false
}

export function useIsMacOS() {
  const { data: platform, isSuccess } = useQuery({
    queryKey: ['platform'],
    queryFn: () => appApi.getPlatform(),
    staleTime: Infinity,
  })
  if (isSuccess) return isDarwinPlatform(platform)
  return typeof navigator !== 'undefined' && /Mac/.test(navigator.platform)
}
