export default {

  login (context, creds) {
    return context.$http.post('/api/v1/login', creds)
      .then((response) => {
        var newSession = {
          'token': response.data.tokenstring,
          'display_name': response.data.display_name,
          'username': response.data.username,
          'jwtexpiry': response.data.jwtexpiry
        }
        window.localStorage.setItem('session', JSON.stringify(newSession))
        context.$store.commit('setSession', newSession)

        // set success to true
        return true
      })
      .catch((error) => {
        if (error) {
          return false
        }
      })
  },

  logout (context) {
    window.localStorage.removeItem('session')
    context.$store.commit('clearSession')
  },

  getSession () {
    let session = JSON.parse(window.localStorage.getItem('session'))
    if (!session) {
      return ''
    }
    return session
  },

  getToken () {
    let session = this.getSession()
    if (!session) {
      return ''
    }
    return session['token']
  },

  getAuthHeader () {
    return {
      'Authorization': 'Bearer ' + this.getToken()
    }
  }
}
