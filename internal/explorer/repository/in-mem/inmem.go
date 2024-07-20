package inmem

import "github.com/noelukwa/git-explorer/internal/explorer/repository"

type repositoryFactory struct{}

func NewRepositoryFactory() repository.RepositoryFactory {
	return &repositoryFactory{}
}

func (f *repositoryFactory) IntentRepository() repository.IntentRepository {
	return newIntentRepository()
}

func (f *repositoryFactory) RemoteRepository() repository.RemoteRepository {
	return newGitRemoteRepository()
}
