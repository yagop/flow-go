package marketplace

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	nbaContract "github.com/dapperlabs/nba-smart-contracts/lib/go/contracts"
	nbaTemplates "github.com/dapperlabs/nba-smart-contracts/lib/go/templates"
	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	coreContract "github.com/onflow/flow-nft/lib/go/contracts"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	"github.com/onflow/flow-go-sdk/client"
	"github.com/onflow/flow-go-sdk/crypto"
)

// MarketPlaceSimulator simulates continuous variable load with interactions
type MarketPlaceSimulator struct {
	log               zerolog.Logger
	networkConfig     *NetworkConfig
	simulatorConfig   *SimulatorConfig
	nbaTopshotAccount *flowAccount
	marketAccounts    []marketPlaceAccount
	availableAccounts chan *marketPlaceAccount
	stopped           bool
	flowClient        *client.Client
	txTracker         *TxTracker
}

func NewMarketPlaceSimulator(
	log zerolog.Logger,
	networkConfig *NetworkConfig,
	simulatorConfig *SimulatorConfig,
) *MarketPlaceSimulator {
	sim := &MarketPlaceSimulator{
		log:               log,
		networkConfig:     networkConfig,
		simulatorConfig:   simulatorConfig,
		marketAccounts:    make([]marketPlaceAccount, 0),
		availableAccounts: make(chan *marketPlaceAccount, simulatorConfig.NumberOfAccounts),
		stopped:           false,
	}

	err := sim.Setup()
	if err != nil {
		panic(err)
	}
	return sim
}

func (m *MarketPlaceSimulator) Setup() error {

	var err error
	// setup client
	m.flowClient, err = client.New(m.networkConfig.AccessNodeAddresses[0], grpc.WithInsecure())
	if err != nil {
		return nil
	}

	// setup tracker (TODO simplify this by using default values and empty txStatsTracker)
	m.txTracker, err = NewTxTracker(m.log, 1000, 10, m.networkConfig.AccessNodeAddresses[0], 1)
	if err != nil {
		return nil
	}

	// load service account
	serviceAcc, err := loadServiceAccount(m.flowClient,
		m.networkConfig.ServiceAccountAddress,
		m.networkConfig.ServiceAccountPrivateKeyHex)
	if err != nil {
		return fmt.Errorf("error loading service account %w", err)
	}

	accounts, err := m.createAccounts(serviceAcc, m.simulatorConfig.NumberOfAccounts+1) // first one is for nba
	if err != nil {
		return err
	}

	// set the nbatopshot account first
	m.nbaTopshotAccount = &accounts[0]
	m.simulatorConfig.NBATopshotAddress = accounts[0].Address()
	accounts = accounts[1:]

	// setup and deploy contracts
	err = m.setupContracts()
	if err != nil {
		return err
	}

	// mint moments
	err = m.mintMoments()
	if err != nil {
		return err
	}

	// setup marketplace accounts
	err = m.setupMarketplaceAccounts(accounts)
	if err != nil {
		return err
	}

	// distribute moments

	return nil
}

func (m *MarketPlaceSimulator) setupContracts() error {

	// deploy nonFungibleContract
	err := m.deployContract("NonFungibleToken", coreContract.NonFungibleToken())
	if err != nil {
		return err
	}

	err = m.deployContract("TopShot", nbaContract.GenerateTopShotContract(m.nbaTopshotAccount.Address().Hex()))
	if err != nil {
		return err
	}

	err = m.deployContract("TopShotShardedCollection", nbaContract.GenerateTopShotShardedCollectionContract(m.nbaTopshotAccount.Address().Hex(),
		m.nbaTopshotAccount.Address().Hex()))
	if err != nil {
		return err
	}

	err = m.deployContract("TopshotAdminReceiver", nbaContract.GenerateTopshotAdminReceiverContract(m.nbaTopshotAccount.Address().Hex(),
		m.nbaTopshotAccount.Address().Hex()))
	if err != nil {
		return err
	}

	err = m.deployContract("Market", nbaContract.GenerateTopShotMarketContract(m.networkConfig.FungibleTokenAddress.Hex(),
		m.nbaTopshotAccount.Address().Hex(),
		m.nbaTopshotAccount.Address().Hex()))

	return err
}

func (m *MarketPlaceSimulator) mintMoments() error {
	blockRef, err := m.flowClient.GetLatestBlockHeader(context.Background(), false)
	if err != nil {
		return err
	}

	nbaAddress := m.nbaTopshotAccount.Address()

	// numBuckets := 10
	// // setup nba account to use sharded collections
	// script := nbaTemplates.GenerateSetupShardedCollectionScript(*nbaAddress, *nbaAddress, numBuckets)
	// tx := flowsdk.NewTransaction().
	// 	SetReferenceBlockID(blockRef.ID).
	// 	SetScript(script)

	// result, err := m.sendTxAndWait(tx, m.nbaTopshotAccount)

	// if err != nil || result.Error != nil {
	// 	m.log.Error().Msgf("error setting up the nba account to us sharded collections: %w , %w", result.Error, err)
	// 	return err
	// }

	// TODO add many plays and add many sets (adds plays to sets)
	// this adds a play with id 0
	script := nbaTemplates.GenerateMintPlayScript(*nbaAddress, *samplePlay())
	tx := flowsdk.NewTransaction().
		SetReferenceBlockID(blockRef.ID).
		SetScript(script)

	result, err := m.sendTxAndWait(tx, m.nbaTopshotAccount)

	if err != nil || result.Error != nil {
		m.log.Error().Msgf("minting a play failed: %w , %w", result.Error, err)
		return err
	}

	m.log.Info().Msgf("a play has been minted")

	// this creates set with id 0
	script = nbaTemplates.GenerateMintSetScript(*nbaAddress, "test set")
	tx = flowsdk.NewTransaction().
		SetReferenceBlockID(blockRef.ID).
		SetScript(script)

	result, err = m.sendTxAndWait(tx, m.nbaTopshotAccount)

	if err != nil || result.Error != nil {
		m.log.Error().Msgf("minting a set failed: %w , %w", result.Error, err)
		return err
	}

	m.log.Info().Msgf("a set has been minted")

	script = nbaTemplates.GenerateAddPlaysToSetScript(*nbaAddress, 1, []uint32{1})
	tx = flowsdk.NewTransaction().
		SetReferenceBlockID(blockRef.ID).
		SetScript(script)

	result, err = m.sendTxAndWait(tx, m.nbaTopshotAccount)

	if err != nil || result.Error != nil {
		m.log.Error().Msgf("adding a play to a set has been failed: %w , %w", result.Error, err)
		return err
	}

	m.log.Info().Msgf("play added to a set")

	// batchSize := 100
	// steps := m.simulatorConfig.NumberOfMoments / batchSize
	// totalMinted := 0
	// for i := 0; i < steps; i++ {
	// 	// mint a lot of moments
	// 	script = nbaTemplates.GenerateBatchMintMomentScript(*nbaAddress, *nbaAddress, 1, 1, uint64(batchSize))
	// 	tx = flowsdk.NewTransaction().
	// 		SetReferenceBlockID(blockRef.ID).
	// 		SetScript(script)

	// 	result, err = m.sendTxAndWait(tx, m.nbaTopshotAccount)
	// 	if err != nil || result.Error != nil {
	// 		m.log.Error().Msgf("adding a play to a set has been failed: %w , %w", result.Error, err)
	// 		return err
	// 	}
	// 	totalMinted += batchSize
	// }

	// m.log.Info().Msgf("%d moment has been minted", totalMinted)
	return nil
}

func (m *MarketPlaceSimulator) setupMarketplaceAccounts(accounts []flowAccount) error {
	// setup marketplace accounts
	// break accounts into batches of 10
	// TODO not share the same client
	blockRef, err := m.flowClient.GetLatestBlockHeader(context.Background(), false)
	if err != nil {
		return err
	}
	groupSize := 10
	numBuckets := 10
	// momentCounter := uint64(1)
	totalMinted := 0
	batchSize := 100

	for i := 0; i < len(accounts); i += groupSize {
		group := accounts[i : i+groupSize]
		// randomly select an access nodes
		n := len(m.networkConfig.AccessNodeAddresses)
		accessNode := m.networkConfig.AccessNodeAddresses[rand.Intn(n)]

		for _, acc := range group {
			ma := newMarketPlaceAccount(&acc, group, m.log, m.simulatorConfig, accessNode)
			m.marketAccounts = append(m.marketAccounts, *ma)
			m.availableAccounts <- ma
			// setup account to be able to intract with nba

			script := nbaTemplates.GenerateSetupShardedCollectionScript(*m.nbaTopshotAccount.Address(), *m.nbaTopshotAccount.Address(), numBuckets)
			tx := flowsdk.NewTransaction().
				SetReferenceBlockID(blockRef.ID).
				SetScript(script)

			result, err := m.sendTxAndWait(tx, ma.Account())
			fmt.Println(">>e>", err)
			fmt.Println(">>r>", result)

			// mint moments for that account
			script = nbaTemplates.GenerateBatchMintMomentScript(*m.nbaTopshotAccount.Address(), *ma.Account().Address(), 1, 1, uint64(batchSize))
			tx = flowsdk.NewTransaction().
				SetReferenceBlockID(blockRef.ID).
				SetScript(script)

			result, err = m.sendTxAndWait(tx, m.nbaTopshotAccount)
			if err != nil || result.Error != nil {
				m.log.Error().Msgf("adding a play to a set has been failed: %w , %w", result.Error, err)
				return err
			}
			fmt.Println(">>e>", err)
			fmt.Println(">>r>", result)
			totalMinted += batchSize

			// get moments
			ma.GetMoments()

			// TODO RAMTIN switch me with GenerateFulfillPackScript
			// //  transfer some moments
			// moments := []uint64{momentCounter, momentCounter + 1, momentCounter + 2, momentCounter + 3, momentCounter + 4}
			// // script = nbaTemplates.GenerateFulfillPackScript(*m.nbaTopshotAccount.Address(), *m.nbaTopshotAccount.Address(), *ma.Account().Address(), moments)
			// script = nbaTemplates.GenerateBatchTransferMomentScript(*m.nbaTopshotAccount.Address(), *m.nbaTopshotAccount.Address(), *ma.Account().Address(), moments)
			// tx = flowsdk.NewTransaction().
			// 	SetReferenceBlockID(blockRef.ID).
			// 	SetScript(script)

			// result, err = m.sendTxAndWait(tx, m.nbaTopshotAccount)
			// fmt.Println("2>>e>", err)
			// fmt.Println("2>>r>", result)
			// momentCounter += 5
		}
	}

	return nil
}

func (m *MarketPlaceSimulator) Run() error {

	// acc := <-lg.availableAccounts
	// defer func() { lg.availableAccounts <- acc }()

	// select a random account
	// call Act and put it back to list when is returned
	// go Run (wrap func into a one to return the account back to list)
	return nil
}

func (m *MarketPlaceSimulator) deployContract(name string, contract []byte) error {
	blockRef, err := m.flowClient.GetLatestBlockHeader(context.Background(), false)
	if err != nil {
		return err
	}

	template := `
	transaction {
		prepare(signer: AuthAccount) {
			signer.contracts.add(name: "%s",
			                     code: "%s".decodeHex())
		}
	}
	`

	script := []byte(fmt.Sprintf(template, name, hex.EncodeToString([]byte(contract))))

	deploymentTx := flowsdk.NewTransaction().
		SetReferenceBlockID(blockRef.ID).
		SetScript(script)

	result, err := m.sendTxAndWait(deploymentTx, m.nbaTopshotAccount)

	if err != nil || result.Error != nil {
		m.log.Error().Msgf("contract %s deployment is failed : %w , %w", name, result.Error, err)
	}

	m.log.Info().Msgf("contract %s is deployed : %s", name, result)
	return err
}

// TODO update this to support multiple tx submissions
func (m *MarketPlaceSimulator) sendTxAndWait(tx *flowsdk.Transaction, sender *flowAccount) (*flowsdk.TransactionResult, error) {

	var result *flowsdk.TransactionResult
	var err error

	err = sender.PrepareAndSignTx(tx, 0)
	if err != nil {
		return nil, fmt.Errorf("error preparing and signing the transaction: %w", err)
	}

	err = m.flowClient.SendTransaction(context.Background(), *tx)
	if err != nil {
		return nil, fmt.Errorf("error sending the transaction: %w", err)

	}

	stopped := false
	wg := sync.WaitGroup{}
	m.txTracker.AddTx(tx.ID(),
		nil,
		func(_ flowsdk.Identifier, res *flowsdk.TransactionResult) {
			m.log.Trace().Str("tx_id", tx.ID().String()).Msgf("finalized tx")
			if !stopped {
				stopped = true
				result = res
				wg.Done()
			}
		}, // on finalized
		func(_ flowsdk.Identifier, _ *flowsdk.TransactionResult) {
			m.log.Trace().Str("tx_id", tx.ID().String()).Msgf("sealed tx")
		}, // on sealed
		func(_ flowsdk.Identifier) {
			m.log.Warn().Str("tx_id", tx.ID().String()).Msgf("tx expired")
			if !stopped {
				stopped = true
				wg.Done()
			}
		}, // on expired
		func(_ flowsdk.Identifier) {
			m.log.Warn().Str("tx_id", tx.ID().String()).Msgf("tx timed out")
			if !stopped {
				stopped = true
				wg.Done()
			}
		}, // on timout
		func(_ flowsdk.Identifier, e error) {
			m.log.Error().Err(err).Str("tx_id", tx.ID().String()).Msgf("tx error")
			if !stopped {
				stopped = true
				err = e
				wg.Done()
			}
		}, // on error
		60)
	wg.Add(1)
	wg.Wait()

	return result, nil
}

func (m *MarketPlaceSimulator) createAccounts(serviceAcc *flowAccount, num int) ([]flowAccount, error) {
	m.log.Info().Msgf("creating and funding %d accounts...", num)

	accounts := make([]flowAccount, 0)

	blockRef, err := m.flowClient.GetLatestBlockHeader(context.Background(), false)
	if err != nil {
		return nil, err
	}

	wg := sync.WaitGroup{}

	privKey := randomPrivateKey()
	accountKey := flowsdk.NewAccountKey().
		FromPrivateKey(privKey).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(flowsdk.AccountKeyWeightThreshold)

	// Generate an account creation script
	createAccountTx := flowsdk.NewTransaction().
		SetScript(createAccountsScript(*m.networkConfig.FungibleTokenAddress,
			*m.networkConfig.FlowTokenAddress)).
		SetReferenceBlockID(blockRef.ID).
		SetProposalKey(
			*serviceAcc.address,
			serviceAcc.accountKey.Index,
			serviceAcc.accountKey.SequenceNumber,
		).
		AddAuthorizer(*serviceAcc.address).
		SetPayer(*serviceAcc.address)

	publicKey := bytesToCadenceArray(accountKey.Encode())
	count := cadence.NewInt(num)

	initialTokenAmount, err := cadence.NewUFix64FromParts(
		24*60*60*0.01,
		0,
	)
	if err != nil {
		return nil, err
	}

	err = createAccountTx.AddArgument(publicKey)
	if err != nil {
		return nil, err
	}

	err = createAccountTx.AddArgument(count)
	if err != nil {
		return nil, err
	}

	err = createAccountTx.AddArgument(initialTokenAmount)
	if err != nil {
		return nil, err
	}

	// TODO replace with account.Sign
	serviceAcc.signerLock.Lock()
	err = createAccountTx.SignEnvelope(
		*serviceAcc.address,
		serviceAcc.accountKey.Index,
		serviceAcc.signer,
	)
	if err != nil {
		return nil, err
	}
	serviceAcc.accountKey.SequenceNumber++
	serviceAcc.signerLock.Unlock()

	err = m.flowClient.SendTransaction(context.Background(), *createAccountTx)
	if err != nil {
		return nil, err
	}

	wg.Add(1)

	i := 0

	m.txTracker.AddTx(createAccountTx.ID(),
		nil,
		func(_ flowsdk.Identifier, res *flowsdk.TransactionResult) {
			defer wg.Done()

			m.log.Debug().
				Str("status", res.Status.String()).
				Msg("account creation tx executed")

			if res.Error != nil {
				m.log.Error().
					Err(res.Error).
					Msg("account creation tx failed")
			}

			for _, event := range res.Events {
				m.log.Trace().
					Str("event_type", event.Type).
					Str("event", event.String()).
					Msg("account creation tx event")

				if event.Type == flowsdk.EventAccountCreated {
					accountCreatedEvent := flowsdk.AccountCreatedEvent(event)
					accountAddress := accountCreatedEvent.Address()

					m.log.Debug().
						Hex("address", accountAddress.Bytes()).
						Msg("new account created")

					signer := crypto.NewInMemorySigner(privKey, accountKey.HashAlgo)

					newAcc := newFlowAccount(i, &accountAddress, accountKey, signer)
					i++

					accounts = append(accounts, *newAcc)

					m.log.Debug().
						Hex("address", accountAddress.Bytes()).
						Msg("new account added")
				}
			}
		},
		nil, // on sealed
		func(_ flowsdk.Identifier) {
			m.log.Error().Msg("setup transaction (account creation) has expired")
			wg.Done()
		}, // on expired
		func(_ flowsdk.Identifier) {
			m.log.Error().Msg("setup transaction (account creation) has timed out")
			wg.Done()
		}, // on timeout
		func(_ flowsdk.Identifier, err error) {
			m.log.Error().Err(err).Msg("setup transaction (account creation) encountered an error")
			wg.Done()
		}, // on error
		120)

	wg.Wait()

	m.log.Info().Msgf("created %d accounts", len(accounts))

	return accounts, nil
}

type marketPlaceAccount struct {
	log             zerolog.Logger
	account         *flowAccount
	friends         []flowAccount
	flowClient      *client.Client
	txTracker       *TxTracker
	simulatorConfig *SimulatorConfig
}

func newMarketPlaceAccount(account *flowAccount,
	friends []flowAccount,
	log zerolog.Logger,
	simulatorConfig *SimulatorConfig,
	accessNodeAddr string) *marketPlaceAccount {
	txTracker, err := NewTxTracker(log,
		10, // max in flight transactions
		1,  // number of workers
		accessNodeAddr,
		1, // number of accounts
	)
	if err != nil {
		panic(err)
	}
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator

	fclient, err := client.New(accessNodeAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	return &marketPlaceAccount{
		log:             log,
		account:         account,
		friends:         friends,
		txTracker:       txTracker,
		flowClient:      fclient,
		simulatorConfig: simulatorConfig,
	}
}

func (m *marketPlaceAccount) Account() *flowAccount {
	return m.account
}

// let collectionRef = acct.getCapability(/public/ShardedMomentCollection)
// .borrow<&{TopShot.MomentCollectionPublic}>()!

func (m *marketPlaceAccount) GetMoments() []uint {

	template := `
	import TopShot from 0x%s

	pub fun main(account: Address): [UInt64] {

		let acct = getAccount(account)

		let collectionRef = acct.borrow<&TopShotShardedCollection.ShardedCollection>(from: /storage/ShardedMomentCollection)!

		log(collectionRef.getIDs())

		return collectionRef.getIDs()
	}
	`

	script := []byte(fmt.Sprintf(template, m.simulatorConfig.NBATopshotAddress.String()))

	res, err := m.flowClient.ExecuteScriptAtLatestBlock(context.Background(), script, []cadence.Value{cadence.Address(*m.account.Address())})

	fmt.Println(">>>>>", res)
	fmt.Println(">>>>>", err)
	return nil
}

func (m *marketPlaceAccount) Act() {

	// nbaTemplates.GenerateTransferMomentScript(nbaTopshotAddress, nbaTopshotAddress, recipientAddr flow.Address, tokenID int)

	// with some chance don't do anything

	// query for active listings to buy

	// // randomly select one or two friend and send assets
	// assets := m.GetAssets()

	// assetToMove := assets[rand.Intn(len(assets))]

	// _ = assetToMove

	// // TODO txScript for assetToMove
	// txScript := []byte("")

	// tx := flowsdk.NewTransaction().
	// 	SetReferenceBlockID(blockRef).
	// 	SetScript(txScript).
	// 	SetProposalKey(*m.account.address, 0, m.account.seqNumber).
	// 	SetPayer(*m.account.address).
	// 	AddAuthorizer(*m.account.address)

	// err = m.account.signTx(tx, 0)
	// if err != nil {
	// 	m.log.Error().Err(err).Msgf("error signing transaction")
	// 	return
	// }

	// // wait till success and then update the list
	// // send tx
	// err = m.flowClient.SendTransaction(context.Background(), *tx)
	// if err != nil {
	// 	m.log.Error().Err(err).Msgf("error sending transaction")
	// 	return
	// }

	// // tracking
	// stopped := false
	// wg := sync.WaitGroup{}
	// m.txTracker.AddTx(tx.ID(),
	// 	nil,
	// 	func(_ flowsdk.Identifier, res *flowsdk.TransactionResult) {
	// 		m.log.Trace().Str("tx_id", tx.ID().String()).Msgf("finalized tx")
	// 	}, // on finalized
	// 	func(_ flowsdk.Identifier, _ *flowsdk.TransactionResult) {
	// 		m.log.Trace().Str("tx_id", tx.ID().String()).Msgf("sealed tx")
	// 		if !stopped {
	// 			stopped = true
	// 			wg.Done()
	// 		}
	// 	}, // on sealed
	// 	func(_ flowsdk.Identifier) {
	// 		m.log.Warn().Str("tx_id", tx.ID().String()).Msgf("tx expired")
	// 		if !stopped {
	// 			stopped = true
	// 			wg.Done()
	// 		}
	// 	}, // on expired
	// 	func(_ flowsdk.Identifier) {
	// 		m.log.Warn().Str("tx_id", tx.ID().String()).Msgf("tx timed out")
	// 		if !stopped {
	// 			stopped = true
	// 			wg.Done()
	// 		}
	// 	}, // on timout
	// 	func(_ flowsdk.Identifier, err error) {
	// 		m.log.Error().Err(err).Str("tx_id", tx.ID().String()).Msgf("tx error")
	// 		if !stopped {
	// 			stopped = true
	// 			wg.Done()
	// 		}
	// 	}, // on error
	// 	60)
	// wg.Add(1)
	// wg.Wait()

	// return
}