// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package matrix_ai

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// PlaceOrder is the `placeOrder` instruction.
type PlaceOrder struct {
	OrderId  *[16]uint8
	Duration *uint32
	Metadata *string

	// [0] = [WRITE] machine
	//
	// [1] = [WRITE] order
	//
	// [2] = [WRITE, SIGNER] buyer
	//
	// [3] = [WRITE] buyerAta
	//
	// [4] = [WRITE] vault
	//
	// [5] = [] mint
	//
	// [6] = [] tokenProgram
	//
	// [7] = [] associatedTokenProgram
	//
	// [8] = [] systemProgram
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewPlaceOrderInstructionBuilder creates a new `PlaceOrder` instruction builder.
func NewPlaceOrderInstructionBuilder() *PlaceOrder {
	nd := &PlaceOrder{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 9),
	}
	return nd
}

// SetOrderId sets the "orderId" parameter.
func (inst *PlaceOrder) SetOrderId(orderId [16]uint8) *PlaceOrder {
	inst.OrderId = &orderId
	return inst
}

// SetDuration sets the "duration" parameter.
func (inst *PlaceOrder) SetDuration(duration uint32) *PlaceOrder {
	inst.Duration = &duration
	return inst
}

// SetMetadata sets the "metadata" parameter.
func (inst *PlaceOrder) SetMetadata(metadata string) *PlaceOrder {
	inst.Metadata = &metadata
	return inst
}

// SetMachineAccount sets the "machine" account.
func (inst *PlaceOrder) SetMachineAccount(machine ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(machine).WRITE()
	return inst
}

// GetMachineAccount gets the "machine" account.
func (inst *PlaceOrder) GetMachineAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetOrderAccount sets the "order" account.
func (inst *PlaceOrder) SetOrderAccount(order ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(order).WRITE()
	return inst
}

// GetOrderAccount gets the "order" account.
func (inst *PlaceOrder) GetOrderAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

// SetBuyerAccount sets the "buyer" account.
func (inst *PlaceOrder) SetBuyerAccount(buyer ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[2] = ag_solanago.Meta(buyer).WRITE().SIGNER()
	return inst
}

// GetBuyerAccount gets the "buyer" account.
func (inst *PlaceOrder) GetBuyerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(2)
}

// SetBuyerAtaAccount sets the "buyerAta" account.
func (inst *PlaceOrder) SetBuyerAtaAccount(buyerAta ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[3] = ag_solanago.Meta(buyerAta).WRITE()
	return inst
}

// GetBuyerAtaAccount gets the "buyerAta" account.
func (inst *PlaceOrder) GetBuyerAtaAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(3)
}

// SetVaultAccount sets the "vault" account.
func (inst *PlaceOrder) SetVaultAccount(vault ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[4] = ag_solanago.Meta(vault).WRITE()
	return inst
}

// GetVaultAccount gets the "vault" account.
func (inst *PlaceOrder) GetVaultAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(4)
}

// SetMintAccount sets the "mint" account.
func (inst *PlaceOrder) SetMintAccount(mint ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[5] = ag_solanago.Meta(mint)
	return inst
}

// GetMintAccount gets the "mint" account.
func (inst *PlaceOrder) GetMintAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(5)
}

// SetTokenProgramAccount sets the "tokenProgram" account.
func (inst *PlaceOrder) SetTokenProgramAccount(tokenProgram ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[6] = ag_solanago.Meta(tokenProgram)
	return inst
}

// GetTokenProgramAccount gets the "tokenProgram" account.
func (inst *PlaceOrder) GetTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(6)
}

// SetAssociatedTokenProgramAccount sets the "associatedTokenProgram" account.
func (inst *PlaceOrder) SetAssociatedTokenProgramAccount(associatedTokenProgram ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[7] = ag_solanago.Meta(associatedTokenProgram)
	return inst
}

// GetAssociatedTokenProgramAccount gets the "associatedTokenProgram" account.
func (inst *PlaceOrder) GetAssociatedTokenProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(7)
}

// SetSystemProgramAccount sets the "systemProgram" account.
func (inst *PlaceOrder) SetSystemProgramAccount(systemProgram ag_solanago.PublicKey) *PlaceOrder {
	inst.AccountMetaSlice[8] = ag_solanago.Meta(systemProgram)
	return inst
}

// GetSystemProgramAccount gets the "systemProgram" account.
func (inst *PlaceOrder) GetSystemProgramAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(8)
}

func (inst PlaceOrder) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_PlaceOrder,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst PlaceOrder) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *PlaceOrder) Validate() error {
	// Check whether all (required) parameters are set:
	{
		if inst.OrderId == nil {
			return errors.New("OrderId parameter is not set")
		}
		if inst.Duration == nil {
			return errors.New("Duration parameter is not set")
		}
		if inst.Metadata == nil {
			return errors.New("Metadata parameter is not set")
		}
	}

	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Machine is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Order is not set")
		}
		if inst.AccountMetaSlice[2] == nil {
			return errors.New("accounts.Buyer is not set")
		}
		if inst.AccountMetaSlice[3] == nil {
			return errors.New("accounts.BuyerAta is not set")
		}
		if inst.AccountMetaSlice[4] == nil {
			return errors.New("accounts.Vault is not set")
		}
		if inst.AccountMetaSlice[5] == nil {
			return errors.New("accounts.Mint is not set")
		}
		if inst.AccountMetaSlice[6] == nil {
			return errors.New("accounts.TokenProgram is not set")
		}
		if inst.AccountMetaSlice[7] == nil {
			return errors.New("accounts.AssociatedTokenProgram is not set")
		}
		if inst.AccountMetaSlice[8] == nil {
			return errors.New("accounts.SystemProgram is not set")
		}
	}
	return nil
}

func (inst *PlaceOrder) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("PlaceOrder")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=3]").ParentFunc(func(paramsBranch ag_treeout.Branches) {
						paramsBranch.Child(ag_format.Param(" OrderId", *inst.OrderId))
						paramsBranch.Child(ag_format.Param("Duration", *inst.Duration))
						paramsBranch.Child(ag_format.Param("Metadata", *inst.Metadata))
					})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=9]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("               machine", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("                 order", inst.AccountMetaSlice.Get(1)))
						accountsBranch.Child(ag_format.Meta("                 buyer", inst.AccountMetaSlice.Get(2)))
						accountsBranch.Child(ag_format.Meta("              buyerAta", inst.AccountMetaSlice.Get(3)))
						accountsBranch.Child(ag_format.Meta("                 vault", inst.AccountMetaSlice.Get(4)))
						accountsBranch.Child(ag_format.Meta("                  mint", inst.AccountMetaSlice.Get(5)))
						accountsBranch.Child(ag_format.Meta("          tokenProgram", inst.AccountMetaSlice.Get(6)))
						accountsBranch.Child(ag_format.Meta("associatedTokenProgram", inst.AccountMetaSlice.Get(7)))
						accountsBranch.Child(ag_format.Meta("         systemProgram", inst.AccountMetaSlice.Get(8)))
					})
				})
		})
}

func (obj PlaceOrder) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	// Serialize `OrderId` param:
	err = encoder.Encode(obj.OrderId)
	if err != nil {
		return err
	}
	// Serialize `Duration` param:
	err = encoder.Encode(obj.Duration)
	if err != nil {
		return err
	}
	// Serialize `Metadata` param:
	err = encoder.Encode(obj.Metadata)
	if err != nil {
		return err
	}
	return nil
}
func (obj *PlaceOrder) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	// Deserialize `OrderId`:
	err = decoder.Decode(&obj.OrderId)
	if err != nil {
		return err
	}
	// Deserialize `Duration`:
	err = decoder.Decode(&obj.Duration)
	if err != nil {
		return err
	}
	// Deserialize `Metadata`:
	err = decoder.Decode(&obj.Metadata)
	if err != nil {
		return err
	}
	return nil
}

// NewPlaceOrderInstruction declares a new PlaceOrder instruction with the provided parameters and accounts.
func NewPlaceOrderInstruction(
	// Parameters:
	orderId [16]uint8,
	duration uint32,
	metadata string,
	// Accounts:
	machine ag_solanago.PublicKey,
	order ag_solanago.PublicKey,
	buyer ag_solanago.PublicKey,
	buyerAta ag_solanago.PublicKey,
	vault ag_solanago.PublicKey,
	mint ag_solanago.PublicKey,
	tokenProgram ag_solanago.PublicKey,
	associatedTokenProgram ag_solanago.PublicKey,
	systemProgram ag_solanago.PublicKey) *PlaceOrder {
	return NewPlaceOrderInstructionBuilder().
		SetOrderId(orderId).
		SetDuration(duration).
		SetMetadata(metadata).
		SetMachineAccount(machine).
		SetOrderAccount(order).
		SetBuyerAccount(buyer).
		SetBuyerAtaAccount(buyerAta).
		SetVaultAccount(vault).
		SetMintAccount(mint).
		SetTokenProgramAccount(tokenProgram).
		SetAssociatedTokenProgramAccount(associatedTokenProgram).
		SetSystemProgramAccount(systemProgram)
}
