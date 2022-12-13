package query

import (
	"context"
	"database/sql"
	errs "errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/zitadel/logging"
	"golang.org/x/text/language"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/projection"
	projection_old "github.com/zitadel/zitadel/internal/query/projection"
	"github.com/zitadel/zitadel/internal/telemetry/tracing"
)

type Users struct {
	SearchResponse
	Users []*User
}

type User struct {
	ID                 string
	CreationDate       time.Time
	ChangeDate         time.Time
	ResourceOwner      string
	Sequence           uint64
	State              domain.UserState
	Type               domain.UserType
	Username           string
	LoginNames         database.StringArray
	PreferredLoginName string
	Human              *Human
	Machine            *Machine
}

type Human struct {
	FirstName         string
	LastName          string
	NickName          string
	DisplayName       string
	AvatarKey         string
	PreferredLanguage language.Tag
	Gender            domain.Gender
	Email             string
	IsEmailVerified   bool
	Phone             string
	IsPhoneVerified   bool
}

type Profile struct {
	ID                string
	CreationDate      time.Time
	ChangeDate        time.Time
	ResourceOwner     string
	Sequence          uint64
	FirstName         string
	LastName          string
	NickName          string
	DisplayName       string
	AvatarKey         string
	PreferredLanguage language.Tag
	Gender            domain.Gender
}

type Email struct {
	ID            string
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	Sequence      uint64
	Email         string
	IsVerified    bool
}

type Phone struct {
	ID            string
	CreationDate  time.Time
	ChangeDate    time.Time
	ResourceOwner string
	Sequence      uint64
	Phone         string
	IsVerified    bool
}

type Machine struct {
	Name        string
	Description string
}

type NotifyUser struct {
	ID                 string
	CreationDate       time.Time
	ChangeDate         time.Time
	ResourceOwner      string
	Sequence           uint64
	State              domain.UserState
	Type               domain.UserType
	Username           string
	LoginNames         database.StringArray
	PreferredLoginName string
	FirstName          string
	LastName           string
	NickName           string
	DisplayName        string
	AvatarKey          string
	PreferredLanguage  language.Tag
	Gender             domain.Gender
	LastEmail          string
	VerifiedEmail      string
	LastPhone          string
	VerifiedPhone      string
	PasswordSet        bool
}

type UserSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

var (
	userTable = table{
		name:          projection_old.UserTable,
		instanceIDCol: projection_old.UserInstanceIDCol,
	}
	UserIDCol = Column{
		name:  projection_old.UserIDCol,
		table: userTable,
	}
	UserCreationDateCol = Column{
		name:  projection_old.UserCreationDateCol,
		table: userTable,
	}
	UserChangeDateCol = Column{
		name:  projection_old.UserChangeDateCol,
		table: userTable,
	}
	UserResourceOwnerCol = Column{
		name:  projection_old.UserResourceOwnerCol,
		table: userTable,
	}
	UserInstanceIDCol = Column{
		name:  projection_old.UserInstanceIDCol,
		table: userTable,
	}
	UserStateCol = Column{
		name:  projection_old.UserStateCol,
		table: userTable,
	}
	UserSequenceCol = Column{
		name:  projection_old.UserSequenceCol,
		table: userTable,
	}
	UserUsernameCol = Column{
		name:           projection_old.UserUsernameCol,
		table:          userTable,
		isOrderByLower: true,
	}
	UserTypeCol = Column{
		name:  projection_old.UserTypeCol,
		table: userTable,
	}
	UserOwnerRemovedCol = Column{
		name:  projection_old.UserOwnerRemovedCol,
		table: userTable,
	}

	userLoginNamesTable                 = loginNameTable.setAlias("login_names")
	userLoginNamesUserIDCol             = LoginNameUserIDCol.setTable(userLoginNamesTable)
	userLoginNamesNameCol               = LoginNameNameCol.setTable(userLoginNamesTable)
	userLoginNamesInstanceIDCol         = LoginNameInstanceIDCol.setTable(userLoginNamesTable)
	userLoginNamesOwnerRemovedUserCol   = LoginNameOwnerRemovedUserCol.setTable(userLoginNamesTable)
	userLoginNamesOwnerRemovedPolicyCol = LoginNameOwnerRemovedPolicyCol.setTable(userLoginNamesTable)
	userLoginNamesOwnerRemovedDomainCol = LoginNameOwnerRemovedDomainCol.setTable(userLoginNamesTable)
	userLoginNamesListCol               = Column{
		name:  "loginnames",
		table: userLoginNamesTable,
	}
	userPreferredLoginNameTable                 = loginNameTable.setAlias("preferred_login_name")
	userPreferredLoginNameUserIDCol             = LoginNameUserIDCol.setTable(userPreferredLoginNameTable)
	userPreferredLoginNameCol                   = LoginNameNameCol.setTable(userPreferredLoginNameTable)
	userPreferredLoginNameIsPrimaryCol          = LoginNameIsPrimaryCol.setTable(userPreferredLoginNameTable)
	userPreferredLoginNameInstanceIDCol         = LoginNameInstanceIDCol.setTable(userPreferredLoginNameTable)
	userPreferredLoginNameOwnerRemovedUserCol   = LoginNameOwnerRemovedUserCol.setTable(userPreferredLoginNameTable)
	userPreferredLoginNameOwnerRemovedPolicyCol = LoginNameOwnerRemovedPolicyCol.setTable(userPreferredLoginNameTable)
	userPreferredLoginNameOwnerRemovedDomainCol = LoginNameOwnerRemovedDomainCol.setTable(userPreferredLoginNameTable)
)

var (
	humanTable = table{
		name:          projection_old.UserHumanTable,
		instanceIDCol: projection_old.HumanUserInstanceIDCol,
	}
	// profile
	HumanUserIDCol = Column{
		name:  projection_old.HumanUserIDCol,
		table: humanTable,
	}
	HumanFirstNameCol = Column{
		name:           projection_old.HumanFirstNameCol,
		table:          humanTable,
		isOrderByLower: true,
	}
	HumanLastNameCol = Column{
		name:           projection_old.HumanLastNameCol,
		table:          humanTable,
		isOrderByLower: true,
	}
	HumanNickNameCol = Column{
		name:           projection_old.HumanNickNameCol,
		table:          humanTable,
		isOrderByLower: true,
	}
	HumanDisplayNameCol = Column{
		name:           projection_old.HumanDisplayNameCol,
		table:          humanTable,
		isOrderByLower: true,
	}
	HumanPreferredLanguageCol = Column{
		name:  projection_old.HumanPreferredLanguageCol,
		table: humanTable,
	}
	HumanGenderCol = Column{
		name:  projection_old.HumanGenderCol,
		table: humanTable,
	}
	HumanAvatarURLCol = Column{
		name:  projection_old.HumanAvatarURLCol,
		table: humanTable,
	}

	// email
	HumanEmailCol = Column{
		name:           projection_old.HumanEmailCol,
		table:          humanTable,
		isOrderByLower: true,
	}
	HumanIsEmailVerifiedCol = Column{
		name:  projection_old.HumanIsEmailVerifiedCol,
		table: humanTable,
	}

	// phone
	HumanPhoneCol = Column{
		name:  projection_old.HumanPhoneCol,
		table: humanTable,
	}
	HumanIsPhoneVerifiedCol = Column{
		name:  projection_old.HumanIsPhoneVerifiedCol,
		table: humanTable,
	}
)

var (
	machineTable = table{
		name:          projection_old.UserMachineTable,
		instanceIDCol: projection_old.MachineUserInstanceIDCol,
	}
	MachineUserIDCol = Column{
		name:  projection_old.MachineUserIDCol,
		table: machineTable,
	}
	MachineNameCol = Column{
		name:           projection_old.MachineNameCol,
		table:          machineTable,
		isOrderByLower: true,
	}
	MachineDescriptionCol = Column{
		name:  projection_old.MachineDescriptionCol,
		table: machineTable,
	}
)

var (
	notifyTable = table{
		name:          projection_old.UserNotifyTable,
		instanceIDCol: projection_old.NotifyInstanceIDCol,
	}
	NotifyUserIDCol = Column{
		name:  projection_old.NotifyUserIDCol,
		table: notifyTable,
	}
	NotifyEmailCol = Column{
		name:           projection_old.NotifyLastEmailCol,
		table:          notifyTable,
		isOrderByLower: true,
	}
	NotifyVerifiedEmailCol = Column{
		name:           projection_old.NotifyVerifiedEmailCol,
		table:          notifyTable,
		isOrderByLower: true,
	}
	NotifyPhoneCol = Column{
		name:  projection_old.NotifyLastPhoneCol,
		table: notifyTable,
	}
	NotifyVerifiedPhoneCol = Column{
		name:  projection_old.NotifyVerifiedPhoneCol,
		table: notifyTable,
	}
	NotifyPasswordSetCol = Column{
		name:  projection_old.NotifyPasswordSetCol,
		table: notifyTable,
	}
)

func addUserWithoutOwnerRemoved(eq map[string]interface{}) {
	eq[UserOwnerRemovedCol.identifier()] = false
	eq[userLoginNamesOwnerRemovedUserCol.identifier()] = false
	eq[userLoginNamesOwnerRemovedPolicyCol.identifier()] = false
	eq[userLoginNamesOwnerRemovedDomainCol.identifier()] = false
	eq[userPreferredLoginNameOwnerRemovedUserCol.identifier()] = false
	eq[userPreferredLoginNameOwnerRemovedPolicyCol.identifier()] = false
	eq[userPreferredLoginNameOwnerRemovedDomainCol.identifier()] = false
}

func (q *Queries) GetUserByID(ctx context.Context, shouldTriggerBulk bool, userID string, withOwnerRemoved bool, queries ...SearchQuery) (_ *User, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	user := projection.NewUserWithOwner(userID, authz.GetInstance(ctx).InstanceID(), authz.GetCtxData(ctx).ResourceOwner)
	events, err := q.eventstore.Filter(ctx, user.SearchQuery(ctx))
	if err != nil {
		return nil, err
	}
	user.Reduce(events)
	if len(events) == 0 || (!withOwnerRemoved && user.OwnerRemoved) {
		return nil, errors.ThrowNotFound(nil, "QUERY-A6ZyK", "Errors.User.NotFound")
	}

	loginNames := projection.NewLoginNamesWithOwner(user.ID, authz.GetInstance(ctx).InstanceID(), user.ResourceOwner)
	events, err = q.eventstore.Filter(ctx, loginNames.SearchQuery(ctx))
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-FBg21", "Errors.Query.SQLStatment")
	}
	loginNames.Reduce(events)

	return mapUser(user, loginNames), nil
}

func mapUser(user *projection.User, loginNames *projection.LoginNames) *User {
	u := &User{
		ID:            user.ID,
		CreationDate:  user.CreationDate,
		ChangeDate:    user.ChangeDate,
		ResourceOwner: user.ResourceOwner,
		Sequence:      user.Sequence,
		State:         user.State,
		Type:          user.Type,
		Username:      user.Username,
	}

	if user.Human != nil {
		u.Human = new(Human)
		u.Human.FirstName = user.Human.Profile.FirstName
		u.Human.LastName = user.Human.Profile.LastName
		u.Human.NickName = user.Human.Profile.NickName
		u.Human.DisplayName = user.Human.Profile.DisplayName
		u.Human.AvatarKey = user.Human.Profile.AvatarKey
		u.Human.PreferredLanguage = user.Human.Profile.PreferredLanguage
		u.Human.Gender = user.Human.Profile.Gender
		u.Human.Email = user.Human.Email.Address
		u.Human.IsEmailVerified = user.Human.Email.IsVerified
		u.Human.Phone = user.Human.Phone.Number
		u.Human.IsPhoneVerified = user.Human.Phone.IsVerified
	} else if user.Machine != nil {
		u.Machine = new(Machine)
		u.Machine.Description = user.Machine.Description
		u.Machine.Name = user.Machine.Name
	}

	for _, loginName := range loginNames.LoginNames {
		u.LoginNames = append(u.LoginNames, loginName.Name)
		if loginName.IsPrimary {
			u.PreferredLoginName = loginName.Name
		}
	}

	return u
}

func (q *Queries) GetUser(ctx context.Context, shouldTriggerBulk bool, withResourceOwnerRemoved bool, queries ...SearchQuery) (_ *User, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	if shouldTriggerBulk {
		logging.OnError(
			projection_old.UserProjection.Trigger(ctx),
		).Debug("unable to trigger")
		logging.OnError(
			projection_old.LoginNameProjection.Trigger(ctx),
		).Debug("unable to trigger")
	}

	query, scan := prepareUserQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{
		UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withResourceOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Dnhr2", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) GetHumanProfile(ctx context.Context, userID string, withOwnerRemoved bool, queries ...SearchQuery) (_ *Profile, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareProfileQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{
		UserIDCol.identifier():         userID,
		UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Dgbg2", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) GetHumanEmail(ctx context.Context, userID string, withOwnerRemoved bool, queries ...SearchQuery) (_ *Email, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareEmailQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{
		UserIDCol.identifier():         userID,
		UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-BHhj3", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) GetHumanPhone(ctx context.Context, userID string, withOwnerRemoved bool, queries ...SearchQuery) (_ *Phone, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := preparePhoneQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{
		UserIDCol.identifier():         userID,
		UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Dg43g", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) GetNotifyUserByID(ctx context.Context, shouldTriggered bool, userID string, withOwnerRemoved bool, queries ...SearchQuery) (_ *NotifyUser, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	if shouldTriggered {
		logging.OnError(
			projection_old.UserProjection.Trigger(ctx),
		).Debug("unable to trigger")
		logging.OnError(
			projection_old.LoginNameProjection.Trigger(ctx),
		).Debug("unable to trigger")
	}

	query, scan := prepareNotifyUserQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{
		UserIDCol.identifier():         userID,
		UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Err3g", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) GetNotifyUser(ctx context.Context, shouldTriggered bool, withOwnerRemoved bool, queries ...SearchQuery) (_ *NotifyUser, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	if shouldTriggered {
		logging.OnError(
			projection_old.UserProjection.Trigger(ctx),
		).Debug("unable to trigger")
		logging.OnError(
			projection_old.LoginNameProjection.Trigger(ctx),
		).Debug("unable to trigger")
	}

	query, scan := prepareNotifyUserQuery()
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{
		UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID(),
	}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Err3g", "Errors.Query.SQLStatment")
	}

	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *Queries) SearchUsers(ctx context.Context, queries *UserSearchQueries, withOwnerRemoved bool) (_ *Users, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareUsersQuery()
	eq := sq.Eq{UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID()}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := queries.toQuery(query).Where(eq).
		ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Dgbg2", "Errors.Query.SQLStatment")
	}

	rows, err := q.client.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-AG4gs", "Errors.Internal")
	}
	users, err := scan(rows)
	if err != nil {
		return nil, err
	}
	users.LatestSequence, err = q.latestSequence(ctx, userTable)
	return users, err
}

func (q *Queries) IsUserUnique(ctx context.Context, username, email, resourceOwner string, withOwnerRemoved bool) (_ bool, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareUserUniqueQuery()
	queries := make([]SearchQuery, 0, 3)
	if username != "" {
		usernameQuery, err := NewUserUsernameSearchQuery(username, TextEquals)
		if err != nil {
			return false, err
		}
		queries = append(queries, usernameQuery)
	}
	if email != "" {
		emailQuery, err := NewUserEmailSearchQuery(email, TextEquals)
		if err != nil {
			return false, err
		}
		queries = append(queries, emailQuery)
	}
	if resourceOwner != "" {
		resourceOwnerQuery, err := NewUserResourceOwnerSearchQuery(resourceOwner, TextEquals)
		if err != nil {
			return false, err
		}
		queries = append(queries, resourceOwnerQuery)
	}
	for _, q := range queries {
		query = q.toQuery(query)
	}
	eq := sq.Eq{UserInstanceIDCol.identifier(): authz.GetInstance(ctx).InstanceID()}
	if !withOwnerRemoved {
		addUserWithoutOwnerRemoved(eq)
	}
	stmt, args, err := query.Where(eq).ToSql()
	if err != nil {
		return false, errors.ThrowInternal(err, "QUERY-Dg43g", "Errors.Query.SQLStatment")
	}
	row := q.client.QueryRowContext(ctx, stmt, args...)
	return scan(row)
}

func (q *UserSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func (r *UserSearchQueries) AppendMyResourceOwnerQuery(orgID string) error {
	query, err := NewUserResourceOwnerSearchQuery(orgID, TextEquals)
	if err != nil {
		return err
	}
	r.Queries = append(r.Queries, query)
	return nil
}

func NewUserResourceOwnerSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(UserResourceOwnerCol, value, comparison)
}

func NewUserUsernameSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(UserUsernameCol, value, comparison)
}

func NewUserFirstNameSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(HumanFirstNameCol, value, comparison)
}

func NewUserLastNameSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(HumanLastNameCol, value, comparison)
}

func NewUserNickNameSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(HumanNickNameCol, value, comparison)
}

func NewUserDisplayNameSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(HumanDisplayNameCol, value, comparison)
}

func NewUserEmailSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(HumanEmailCol, value, comparison)
}

func NewUserPhoneSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(HumanPhoneCol, value, comparison)
}

func NewUserVerifiedEmailSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(NotifyVerifiedEmailCol, value, comparison)
}

func NewUserVerifiedPhoneSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(NotifyVerifiedPhoneCol, value, comparison)
}

func NewUserStateSearchQuery(value int32) (SearchQuery, error) {
	return NewNumberQuery(UserStateCol, value, NumberEquals)
}

func NewUserTypeSearchQuery(value int32) (SearchQuery, error) {
	return NewNumberQuery(UserTypeCol, value, NumberEquals)
}

func NewUserPreferredLoginNameSearchQuery(value string, comparison TextComparison) (SearchQuery, error) {
	return NewTextQuery(userPreferredLoginNameCol, value, comparison)
}

func NewUserLoginNamesSearchQuery(value string) (SearchQuery, error) {
	return NewTextQuery(userLoginNamesListCol, value, TextListContains)
}

func NewUserLoginNameExistsQuery(value string, comparison TextComparison) (SearchQuery, error) {
	//linking queries for the subselect
	instanceQuery, err := NewColumnComparisonQuery(LoginNameInstanceIDCol, UserInstanceIDCol, ColumnEquals)
	if err != nil {
		return nil, err
	}
	userIDQuery, err := NewColumnComparisonQuery(LoginNameUserIDCol, UserIDCol, ColumnEquals)
	if err != nil {
		return nil, err
	}
	//text query to select data from the linked sub select
	loginNameQuery, err := NewTextQuery(LoginNameNameCol, value, comparison)
	if err != nil {
		return nil, err
	}
	//full definition of the sub select
	subSelect, err := NewSubSelect(LoginNameUserIDCol, []SearchQuery{instanceQuery, userIDQuery, loginNameQuery})
	if err != nil {
		return nil, err
	}
	// "WHERE * IN (*)" query with subquery as list-data provider
	return NewListQuery(
		UserIDCol,
		subSelect,
		ListIn,
	)
}

func prepareLoginNamesQuery() (string, []interface{}, error) {
	return sq.Select(
		userLoginNamesUserIDCol.identifier(),
		"ARRAY_AGG("+userLoginNamesNameCol.identifier()+")::TEXT[] AS "+userLoginNamesListCol.name,
		userLoginNamesInstanceIDCol.identifier(),
		userLoginNamesOwnerRemovedUserCol.identifier(),
		userLoginNamesOwnerRemovedPolicyCol.identifier(),
		userLoginNamesOwnerRemovedDomainCol.identifier(),
	).From(userLoginNamesTable.identifier()).
		GroupBy(
			userLoginNamesUserIDCol.identifier(),
			userLoginNamesInstanceIDCol.identifier(),
			userLoginNamesOwnerRemovedUserCol.identifier(),
			userLoginNamesOwnerRemovedPolicyCol.identifier(),
			userLoginNamesOwnerRemovedDomainCol.identifier(),
		).ToSql()
}

func preparePreferredLoginNamesQuery() (string, []interface{}, error) {
	return sq.Select(
		userPreferredLoginNameUserIDCol.identifier(),
		userPreferredLoginNameCol.identifier(),
		userPreferredLoginNameInstanceIDCol.identifier(),
		userPreferredLoginNameOwnerRemovedUserCol.identifier(),
		userPreferredLoginNameOwnerRemovedPolicyCol.identifier(),
		userPreferredLoginNameOwnerRemovedDomainCol.identifier(),
	).From(userPreferredLoginNameTable.identifier()).
		Where(sq.Eq{
			userPreferredLoginNameIsPrimaryCol.identifier(): true,
		},
		).ToSql()
}

func prepareUserQuery() (sq.SelectBuilder, func(*sql.Row) (*User, error)) {
	loginNamesQuery, loginNamesArgs, err := prepareLoginNamesQuery()
	if err != nil {
		return sq.SelectBuilder{}, nil
	}
	preferredLoginNameQuery, preferredLoginNameArgs, err := preparePreferredLoginNamesQuery()
	if err != nil {
		return sq.SelectBuilder{}, nil
	}
	return sq.Select(
			UserIDCol.identifier(),
			UserCreationDateCol.identifier(),
			UserChangeDateCol.identifier(),
			UserResourceOwnerCol.identifier(),
			UserSequenceCol.identifier(),
			UserStateCol.identifier(),
			UserTypeCol.identifier(),
			UserUsernameCol.identifier(),
			userLoginNamesListCol.identifier(),
			userPreferredLoginNameCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanFirstNameCol.identifier(),
			HumanLastNameCol.identifier(),
			HumanNickNameCol.identifier(),
			HumanDisplayNameCol.identifier(),
			HumanPreferredLanguageCol.identifier(),
			HumanGenderCol.identifier(),
			HumanAvatarURLCol.identifier(),
			HumanEmailCol.identifier(),
			HumanIsEmailVerifiedCol.identifier(),
			HumanPhoneCol.identifier(),
			HumanIsPhoneVerifiedCol.identifier(),
			MachineUserIDCol.identifier(),
			MachineNameCol.identifier(),
			MachineDescriptionCol.identifier(),
			countColumn.identifier(),
		).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			LeftJoin(join(MachineUserIDCol, UserIDCol)).
			LeftJoin("("+loginNamesQuery+") AS "+userLoginNamesTable.alias+" ON "+
				userLoginNamesUserIDCol.identifier()+" = "+UserIDCol.identifier()+" AND "+
				userLoginNamesInstanceIDCol.identifier()+" = "+UserInstanceIDCol.identifier(),
				loginNamesArgs...).
			LeftJoin("("+preferredLoginNameQuery+") AS "+userPreferredLoginNameTable.alias+" ON "+
				userPreferredLoginNameUserIDCol.identifier()+" = "+UserIDCol.identifier()+" AND "+
				userPreferredLoginNameInstanceIDCol.identifier()+" = "+UserInstanceIDCol.identifier(),
				preferredLoginNameArgs...).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*User, error) {
			u := new(User)
			var count int
			preferredLoginName := sql.NullString{}

			humanID := sql.NullString{}
			firstName := sql.NullString{}
			lastName := sql.NullString{}
			nickName := sql.NullString{}
			displayName := sql.NullString{}
			preferredLanguage := sql.NullString{}
			gender := sql.NullInt32{}
			avatarKey := sql.NullString{}
			email := sql.NullString{}
			isEmailVerified := sql.NullBool{}
			phone := sql.NullString{}
			isPhoneVerified := sql.NullBool{}

			machineID := sql.NullString{}
			name := sql.NullString{}
			description := sql.NullString{}

			err := row.Scan(
				&u.ID,
				&u.CreationDate,
				&u.ChangeDate,
				&u.ResourceOwner,
				&u.Sequence,
				&u.State,
				&u.Type,
				&u.Username,
				&u.LoginNames,
				&preferredLoginName,
				&humanID,
				&firstName,
				&lastName,
				&nickName,
				&displayName,
				&preferredLanguage,
				&gender,
				&avatarKey,
				&email,
				&isEmailVerified,
				&phone,
				&isPhoneVerified,
				&machineID,
				&name,
				&description,
				&count,
			)

			if err != nil || count != 1 {
				if errs.Is(err, sql.ErrNoRows) || count != 1 {
					return nil, errors.ThrowNotFound(err, "QUERY-Dfbg2", "Errors.User.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Bgah2", "Errors.Internal")
			}

			u.PreferredLoginName = preferredLoginName.String

			if humanID.Valid {
				u.Human = &Human{
					FirstName:         firstName.String,
					LastName:          lastName.String,
					NickName:          nickName.String,
					DisplayName:       displayName.String,
					AvatarKey:         avatarKey.String,
					PreferredLanguage: language.Make(preferredLanguage.String),
					Gender:            domain.Gender(gender.Int32),
					Email:             email.String,
					IsEmailVerified:   isEmailVerified.Bool,
					Phone:             phone.String,
					IsPhoneVerified:   isPhoneVerified.Bool,
				}
			} else if machineID.Valid {
				u.Machine = &Machine{
					Name:        name.String,
					Description: description.String,
				}
			}
			return u, nil
		}
}

func prepareProfileQuery() (sq.SelectBuilder, func(*sql.Row) (*Profile, error)) {
	return sq.Select(
			UserIDCol.identifier(),
			UserCreationDateCol.identifier(),
			UserChangeDateCol.identifier(),
			UserResourceOwnerCol.identifier(),
			UserSequenceCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanFirstNameCol.identifier(),
			HumanLastNameCol.identifier(),
			HumanNickNameCol.identifier(),
			HumanDisplayNameCol.identifier(),
			HumanPreferredLanguageCol.identifier(),
			HumanGenderCol.identifier(),
			HumanAvatarURLCol.identifier()).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*Profile, error) {
			p := new(Profile)

			humanID := sql.NullString{}
			firstName := sql.NullString{}
			lastName := sql.NullString{}
			nickName := sql.NullString{}
			displayName := sql.NullString{}
			preferredLanguage := sql.NullString{}
			gender := sql.NullInt32{}
			avatarKey := sql.NullString{}
			err := row.Scan(
				&p.ID,
				&p.CreationDate,
				&p.ChangeDate,
				&p.ResourceOwner,
				&p.Sequence,
				&humanID,
				&firstName,
				&lastName,
				&nickName,
				&displayName,
				&preferredLanguage,
				&gender,
				&avatarKey,
			)
			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, errors.ThrowNotFound(err, "QUERY-HNhb3", "Errors.User.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Rfheq", "Errors.Internal")
			}
			if !humanID.Valid {
				return nil, errors.ThrowPreconditionFailed(nil, "QUERY-WLTce", "Errors.User.NotHuman")
			}

			p.FirstName = firstName.String
			p.LastName = lastName.String
			p.NickName = nickName.String
			p.DisplayName = displayName.String
			p.AvatarKey = avatarKey.String
			p.PreferredLanguage = language.Make(preferredLanguage.String)
			p.Gender = domain.Gender(gender.Int32)

			return p, nil
		}
}

func prepareEmailQuery() (sq.SelectBuilder, func(*sql.Row) (*Email, error)) {
	return sq.Select(
			UserIDCol.identifier(),
			UserCreationDateCol.identifier(),
			UserChangeDateCol.identifier(),
			UserResourceOwnerCol.identifier(),
			UserSequenceCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanEmailCol.identifier(),
			HumanIsEmailVerifiedCol.identifier()).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*Email, error) {
			e := new(Email)

			humanID := sql.NullString{}
			email := sql.NullString{}
			isEmailVerified := sql.NullBool{}

			err := row.Scan(
				&e.ID,
				&e.CreationDate,
				&e.ChangeDate,
				&e.ResourceOwner,
				&e.Sequence,
				&humanID,
				&email,
				&isEmailVerified,
			)
			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, errors.ThrowNotFound(err, "QUERY-Hms2s", "Errors.User.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Nu42d", "Errors.Internal")
			}
			if !humanID.Valid {
				return nil, errors.ThrowPreconditionFailed(nil, "QUERY-pt7HY", "Errors.User.NotHuman")
			}

			e.Email = email.String
			e.IsVerified = isEmailVerified.Bool

			return e, nil
		}
}

func preparePhoneQuery() (sq.SelectBuilder, func(*sql.Row) (*Phone, error)) {
	return sq.Select(
			UserIDCol.identifier(),
			UserCreationDateCol.identifier(),
			UserChangeDateCol.identifier(),
			UserResourceOwnerCol.identifier(),
			UserSequenceCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanPhoneCol.identifier(),
			HumanIsPhoneVerifiedCol.identifier()).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*Phone, error) {
			e := new(Phone)

			humanID := sql.NullString{}
			phone := sql.NullString{}
			isPhoneVerified := sql.NullBool{}

			err := row.Scan(
				&e.ID,
				&e.CreationDate,
				&e.ChangeDate,
				&e.ResourceOwner,
				&e.Sequence,
				&humanID,
				&phone,
				&isPhoneVerified,
			)
			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, errors.ThrowNotFound(err, "QUERY-DAvb3", "Errors.User.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Bmf2h", "Errors.Internal")
			}
			if !humanID.Valid {
				return nil, errors.ThrowPreconditionFailed(nil, "QUERY-hliQl", "Errors.User.NotHuman")
			}

			e.Phone = phone.String
			e.IsVerified = isPhoneVerified.Bool

			return e, nil
		}
}

func prepareNotifyUserQuery() (sq.SelectBuilder, func(*sql.Row) (*NotifyUser, error)) {
	loginNamesQuery, loginNamesArgs, err := prepareLoginNamesQuery()
	if err != nil {
		return sq.SelectBuilder{}, nil
	}
	preferredLoginNameQuery, preferredLoginNameArgs, err := preparePreferredLoginNamesQuery()
	if err != nil {
		return sq.SelectBuilder{}, nil
	}
	return sq.Select(
			UserIDCol.identifier(),
			UserCreationDateCol.identifier(),
			UserChangeDateCol.identifier(),
			UserResourceOwnerCol.identifier(),
			UserSequenceCol.identifier(),
			UserStateCol.identifier(),
			UserTypeCol.identifier(),
			UserUsernameCol.identifier(),
			userLoginNamesListCol.identifier(),
			userPreferredLoginNameCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanFirstNameCol.identifier(),
			HumanLastNameCol.identifier(),
			HumanNickNameCol.identifier(),
			HumanDisplayNameCol.identifier(),
			HumanPreferredLanguageCol.identifier(),
			HumanGenderCol.identifier(),
			HumanAvatarURLCol.identifier(),
			NotifyUserIDCol.identifier(),
			NotifyEmailCol.identifier(),
			NotifyVerifiedEmailCol.identifier(),
			NotifyPhoneCol.identifier(),
			NotifyVerifiedPhoneCol.identifier(),
			NotifyPasswordSetCol.identifier(),
			countColumn.identifier(),
		).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			LeftJoin(join(NotifyUserIDCol, UserIDCol)).
			LeftJoin("("+loginNamesQuery+") AS "+userLoginNamesTable.alias+" ON "+
				userLoginNamesUserIDCol.identifier()+" = "+UserIDCol.identifier()+" AND "+
				userLoginNamesInstanceIDCol.identifier()+" = "+UserInstanceIDCol.identifier(),
				loginNamesArgs...).
			LeftJoin("("+preferredLoginNameQuery+") AS "+userPreferredLoginNameTable.alias+" ON "+
				userPreferredLoginNameUserIDCol.identifier()+" = "+UserIDCol.identifier()+" AND "+
				userPreferredLoginNameInstanceIDCol.identifier()+" = "+UserInstanceIDCol.identifier(),
				preferredLoginNameArgs...).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*NotifyUser, error) {
			u := new(NotifyUser)
			var count int
			loginNames := database.StringArray{}
			preferredLoginName := sql.NullString{}

			humanID := sql.NullString{}
			firstName := sql.NullString{}
			lastName := sql.NullString{}
			nickName := sql.NullString{}
			displayName := sql.NullString{}
			preferredLanguage := sql.NullString{}
			gender := sql.NullInt32{}
			avatarKey := sql.NullString{}

			notifyUserID := sql.NullString{}
			notifyEmail := sql.NullString{}
			notifyVerifiedEmail := sql.NullString{}
			notifyPhone := sql.NullString{}
			notifyVerifiedPhone := sql.NullString{}
			notifyPasswordSet := sql.NullBool{}

			err := row.Scan(
				&u.ID,
				&u.CreationDate,
				&u.ChangeDate,
				&u.ResourceOwner,
				&u.Sequence,
				&u.State,
				&u.Type,
				&u.Username,
				&loginNames,
				&preferredLoginName,
				&humanID,
				&firstName,
				&lastName,
				&nickName,
				&displayName,
				&preferredLanguage,
				&gender,
				&avatarKey,
				&notifyUserID,
				&notifyEmail,
				&notifyVerifiedEmail,
				&notifyPhone,
				&notifyVerifiedPhone,
				&notifyPasswordSet,
				&count,
			)

			if err != nil || count != 1 {
				if errs.Is(err, sql.ErrNoRows) || count != 1 {
					return nil, errors.ThrowNotFound(err, "QUERY-Dgqd2", "Errors.User.NotFound")
				}
				return nil, errors.ThrowInternal(err, "QUERY-Dbwsg", "Errors.Internal")
			}

			if !notifyUserID.Valid {
				return nil, errors.ThrowPreconditionFailed(nil, "QUERY-Sfw3f", "Errors.User.NotFound")
			}

			u.LoginNames = loginNames
			if preferredLoginName.Valid {
				u.PreferredLoginName = preferredLoginName.String
			}
			if humanID.Valid {
				u.FirstName = firstName.String
				u.LastName = lastName.String
				u.NickName = nickName.String
				u.DisplayName = displayName.String
				u.AvatarKey = avatarKey.String
				u.PreferredLanguage = language.Make(preferredLanguage.String)
				u.Gender = domain.Gender(gender.Int32)
			}
			u.LastEmail = notifyEmail.String
			u.VerifiedEmail = notifyVerifiedEmail.String
			u.LastPhone = notifyPhone.String
			u.VerifiedPhone = notifyVerifiedPhone.String
			u.PasswordSet = notifyPasswordSet.Bool

			return u, nil
		}
}

func prepareUserUniqueQuery() (sq.SelectBuilder, func(*sql.Row) (bool, error)) {
	return sq.Select(
			UserIDCol.identifier(),
			UserStateCol.identifier(),
			UserUsernameCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanEmailCol.identifier(),
			HumanIsEmailVerifiedCol.identifier()).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (bool, error) {
			userID := sql.NullString{}
			state := sql.NullInt32{}
			username := sql.NullString{}
			humanID := sql.NullString{}
			email := sql.NullString{}
			isEmailVerified := sql.NullBool{}

			err := row.Scan(
				&userID,
				&state,
				&username,
				&humanID,
				&email,
				&isEmailVerified,
			)
			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return true, nil
				}
				return false, errors.ThrowInternal(err, "QUERY-Cxces", "Errors.Internal")
			}
			return !userID.Valid, nil
		}
}

func prepareUsersQuery() (sq.SelectBuilder, func(*sql.Rows) (*Users, error)) {
	loginNamesQuery, loginNamesArgs, err := prepareLoginNamesQuery()
	if err != nil {
		return sq.SelectBuilder{}, nil
	}
	preferredLoginNameQuery, preferredLoginNameArgs, err := preparePreferredLoginNamesQuery()
	if err != nil {
		return sq.SelectBuilder{}, nil
	}
	return sq.Select(
			UserIDCol.identifier(),
			UserCreationDateCol.identifier(),
			UserChangeDateCol.identifier(),
			UserResourceOwnerCol.identifier(),
			UserSequenceCol.identifier(),
			UserStateCol.identifier(),
			UserTypeCol.identifier(),
			UserUsernameCol.identifier(),
			userLoginNamesListCol.identifier(),
			userPreferredLoginNameCol.identifier(),
			HumanUserIDCol.identifier(),
			HumanFirstNameCol.identifier(),
			HumanLastNameCol.identifier(),
			HumanNickNameCol.identifier(),
			HumanDisplayNameCol.identifier(),
			HumanPreferredLanguageCol.identifier(),
			HumanGenderCol.identifier(),
			HumanAvatarURLCol.identifier(),
			HumanEmailCol.identifier(),
			HumanIsEmailVerifiedCol.identifier(),
			HumanPhoneCol.identifier(),
			HumanIsPhoneVerifiedCol.identifier(),
			MachineUserIDCol.identifier(),
			MachineNameCol.identifier(),
			MachineDescriptionCol.identifier(),
			countColumn.identifier()).
			From(userTable.identifier()).
			LeftJoin(join(HumanUserIDCol, UserIDCol)).
			LeftJoin(join(MachineUserIDCol, UserIDCol)).
			LeftJoin("("+loginNamesQuery+") AS "+userLoginNamesTable.alias+" ON "+
				userLoginNamesUserIDCol.identifier()+" = "+UserIDCol.identifier()+" AND "+
				userLoginNamesInstanceIDCol.identifier()+" = "+UserInstanceIDCol.identifier(),
				loginNamesArgs...).
			LeftJoin("("+preferredLoginNameQuery+") AS "+userPreferredLoginNameTable.alias+" ON "+
				userPreferredLoginNameUserIDCol.identifier()+" = "+UserIDCol.identifier()+" AND "+
				userPreferredLoginNameInstanceIDCol.identifier()+" = "+UserInstanceIDCol.identifier(),
				preferredLoginNameArgs...).
			PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*Users, error) {
			users := make([]*User, 0)
			var count uint64
			for rows.Next() {
				u := new(User)
				loginNames := database.StringArray{}
				preferredLoginName := sql.NullString{}

				humanID := sql.NullString{}
				firstName := sql.NullString{}
				lastName := sql.NullString{}
				nickName := sql.NullString{}
				displayName := sql.NullString{}
				preferredLanguage := sql.NullString{}
				gender := sql.NullInt32{}
				avatarKey := sql.NullString{}
				email := sql.NullString{}
				isEmailVerified := sql.NullBool{}
				phone := sql.NullString{}
				isPhoneVerified := sql.NullBool{}

				machineID := sql.NullString{}
				name := sql.NullString{}
				description := sql.NullString{}

				err := rows.Scan(
					&u.ID,
					&u.CreationDate,
					&u.ChangeDate,
					&u.ResourceOwner,
					&u.Sequence,
					&u.State,
					&u.Type,
					&u.Username,
					&loginNames,
					&preferredLoginName,
					&humanID,
					&firstName,
					&lastName,
					&nickName,
					&displayName,
					&preferredLanguage,
					&gender,
					&avatarKey,
					&email,
					&isEmailVerified,
					&phone,
					&isPhoneVerified,
					&machineID,
					&name,
					&description,
					&count,
				)
				if err != nil {
					return nil, err
				}

				u.LoginNames = loginNames
				if preferredLoginName.Valid {
					u.PreferredLoginName = preferredLoginName.String
				}

				if humanID.Valid {
					u.Human = &Human{
						FirstName:         firstName.String,
						LastName:          lastName.String,
						NickName:          nickName.String,
						DisplayName:       displayName.String,
						AvatarKey:         avatarKey.String,
						PreferredLanguage: language.Make(preferredLanguage.String),
						Gender:            domain.Gender(gender.Int32),
						Email:             email.String,
						IsEmailVerified:   isEmailVerified.Bool,
						Phone:             phone.String,
						IsPhoneVerified:   isPhoneVerified.Bool,
					}
				} else if machineID.Valid {
					u.Machine = &Machine{
						Name:        name.String,
						Description: description.String,
					}
				}

				users = append(users, u)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-frhbd", "Errors.Query.CloseRows")
			}

			return &Users{
				Users: users,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}
