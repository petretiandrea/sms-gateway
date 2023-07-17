package account

type UserAccountService struct {
	repo UserAccountRepository
}

func NewUserAccountService(repository UserAccountRepository) UserAccountService {
	return UserAccountService{repo: repository}
}

type CreateNewAccountParams struct {
	Phone string
}

func (service *UserAccountService) CreateNewAccount(params CreateNewAccountParams) (*UserAccount, error) {
	account := NewUserAccount(params.Phone)
	if _, err := service.repo.Save(account); err == nil {
		return &account, nil
	} else {
		return nil, err
	}
}

func (service *UserAccountService) GetUserAccount(id AccountID) *UserAccount {
	return service.repo.FindById(id)
}

func (service *UserAccountService) GetUserAccountByApiKey(apiKey ApiKey) *UserAccount {
	return service.repo.FindByApiKey(apiKey)
}

func (service *UserAccountService) GetUserAccountApiKey(id AccountID) *ApiKey {
	if account := service.repo.FindById(id); account != nil {
		return &account.ApiKey
	} else {
		return nil
	}
}
