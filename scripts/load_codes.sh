
# demov3 NFT w/ offer: Goat One on demov3
contract=0xb914ad493a0a4fe5a899dc21b66a509bcf8f1ed9

# demov3 NFT w/ offer: Acme Ticket on demov3
contract2=0x2d9729b9f7049bb3cd6c4ed572f7e6f47922ca68

# main/prod eluvio sheep
contract3=0xe70d12af413a3a4caf2e8e182560c7324268b443
contract4=0xd4c8153372b0292b364dac40d0ade37da4c4869a

# prod
#prefix=https://appsvc.svc.eluv.io/main/code-fulfillment
# testing
prefix=http://localhost:2023/dv3/

for i in {1..3}
do
  code=`echo $RANDOM | base64 | tr -d =`

  for c in $contract $contract2 $contract3 $contract4
  do
    curl -s -X POST -H "Content-Type: application/json" \
      -d '{ "url": "https://eluv.io/vouncher-redeem", "codes":  [ "'0$code'" ] }' \
          $prefix/load/$c/0
    curl -s -X POST -H "Content-Type: application/json" \
      -d '{ "url": "https://eluv.io/vouncher-redeem", "codes":  [ "'1$code'" ] }' \
          $prefix/load/$c/1
  done

done
