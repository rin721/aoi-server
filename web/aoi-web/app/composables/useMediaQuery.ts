export function useMediaQuery(query: string) {
  const matches = ref(false)
  let media: MediaQueryList | null = null

  function update() {
    matches.value = Boolean(media?.matches)
  }

  onMounted(() => {
    media = window.matchMedia(query)
    update()
    media.addEventListener("change", update)
  })

  onBeforeUnmount(() => {
    media?.removeEventListener("change", update)
    media = null
  })

  return matches
}


