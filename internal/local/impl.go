package local

type localStore struct {
	userConfig   UserConfigStore
	serverConfig ServerConfigStore
}

func NewLocalStore() LocalStore {
	return &localStore{
		userConfig:   newUserConfigStore(),
		serverConfig: newServerConfigStore(),
	}
}

func (l *localStore) UserConfigStore() UserConfigStore {
	return l.userConfig
}

func (l *localStore) ServerConfigStore() ServerConfigStore {
	return l.serverConfig
}
