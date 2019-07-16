export const setSession = (state, session) => {
  state.session = session
}

export const clearSession = (state) => {
  state.session = null
}

export const appendInterval = (state, interval) => {
  if (!state.intervals) {
    state.intervals = []
  }

  state.intervals.push(interval)
}

export const clearIntervals = (state) => {
  if (state.intervals) {
    for (let i = 0, l = state.intervals.length; i < l; i++) {
      clearInterval(state.intervals[i])
    }
    state.intervals = null
  }
}
